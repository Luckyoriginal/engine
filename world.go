package engine

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	arkecs "github.com/mlange-42/ark/ecs"
)

// World is the interface passed to Setup, systems, and lifecycle hooks.
type World interface {
	// Spawn creates one entity from a pointer to a struct.
	// Each exported field whose type was registered via Comp[T].Register
	// becomes a component on the new entity.
	// Returns the new entity handle.
	Spawn(entity any) arkecs.Entity

	// Despawn removes an entity, triggering OnRemove on its components.
	Despawn(e arkecs.Entity)

	// AddSystems registers systems in the order they will run.
	// Each value must implement Updater, Drawer, or both.
	AddSystems(systems ...any)

	// SetResource stores a singleton value under key.
	SetResource(key, value any)

	// Resource retrieves a singleton value.
	Resource(key any) (any, bool)

	// Ark returns the underlying ark World for direct low-level access.
	Ark() *arkecs.World
}

// ── compEntry ─────────────────────────────────────────────────────────────────

// compEntry holds everything needed to copy and lifecycle-manage one component
// type at entity-spawn time.
type compEntry struct {
	id       arkecs.ID
	typ      reflect.Type
	copyInto func(dst unsafe.Pointer, src reflect.Value)
	// onAdder is non-nil when the component implements OnAdder.
	// Called directly in spawnStruct after data is written, so OnAdd always
	// receives the fully-initialised component value.
	onAdder func(ptr unsafe.Pointer, w World, e arkecs.Entity)
	has     func(e arkecs.Entity) bool
	get     func(e arkecs.Entity) unsafe.Pointer
}

// ── worldImpl ─────────────────────────────────────────────────────────────────

type worldImpl struct {
	ark       arkecs.World
	byType    map[reflect.Type]*compEntry
	systems   []system
	resources map[any]any
}

func newWorldImpl() *worldImpl {
	return &worldImpl{
		ark:       *arkecs.NewWorld(),
		byType:    make(map[reflect.Type]*compEntry),
		resources: make(map[any]any),
	}
}

// ── World interface ───────────────────────────────────────────────────────────

func (w *worldImpl) Ark() *arkecs.World { return &w.ark }

func (w *worldImpl) SetResource(key, value any) { w.resources[key] = value }
func (w *worldImpl) Resource(key any) (any, bool) {
	v, ok := w.resources[key]
	return v, ok
}

type autoField struct {
	offset uintptr
	entry  *compEntry
}

func (w *worldImpl) AddSystems(systems ...any) {
	for _, s := range systems {
		// Auto-build any engine.FilterN / engine.MapN fields.
		buildFilterFields(s, &w.ark)
		// Then call Init if the system wants manual control on top.
		if i, ok := s.(Initer); ok {
			i.Init(w)
		}

		// Auto-detect if this system is an "Auto-Entity System" (has component pointers as fields)
		fields := detectAutoFields(s, w.byType)
		if len(fields) > 0 {
			updater, isUpdate := s.(Updater)
			drawer, isDraw := s.(Drawer)

			var u Updater
			var d Drawer

			// Extract IDs to build the UnsafeFilter
			ids := make([]arkecs.ID, len(fields))
			for i := range fields {
				ids[i] = fields[i].entry.id
			}
			filter := arkecs.NewUnsafeFilter(&w.ark, ids...)

			if isUpdate {
				u = &autoSystemUpdater{world: w, sys: s, fields: fields, filter: filter, u: updater}
			}
			if isDraw {
				d = &autoSystemDrawer{world: w, sys: s, fields: fields, filter: filter, d: drawer}
			}

			w.systems = append(w.systems, system{u, d})
		} else {
			// Standard Global System
			w.systems = append(w.systems, wrapSystem(s))
		}
	}
}

// detectAutoFields checks if a system is an Auto-Entity System by identifying
// exported pointer fields whose element types are registered components.
func detectAutoFields(sys any, byType map[reflect.Type]*compEntry) []autoField {
	v := reflect.ValueOf(sys)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil
	}
	v = v.Elem()
	t := v.Type()

	var fields []autoField
	for i := range v.NumField() {
		structField := t.Field(i)
		if structField.PkgPath != "" {
			continue // Skip unexported fields
		}
		ft := structField.Type
		if ft.Kind() == reflect.Ptr {
			elemType := ft.Elem()
			if entry, ok := byType[elemType]; ok {
				fields = append(fields, autoField{
					offset: structField.Offset,
					entry:  entry,
				})
			}
		}
	}
	return fields
}

// ── Auto-Entity System Wrappers (Zero Reflection per Tick!) ───────────────────

type autoSystemUpdater struct {
	world  *worldImpl
	sys    any
	fields []autoField
	filter arkecs.UnsafeFilter
	u      Updater
}

func (asu *autoSystemUpdater) Update(w World) {
	// Get base memory address of the system struct once
	basePtr := unsafe.Pointer(reflect.ValueOf(asu.sys).Pointer())

	q := asu.filter.Query()
	for q.Next() {
		// Direct Memory Injection: Write component pointers directly to fields using unsafe offsets.
		// Bypasses Go's runtime reflection Set() entirely. 100% compile-speed, zero allocations!
		for i := range asu.fields {
			ptr := q.Get(asu.fields[i].entry.id)
			fieldPtr := unsafe.Pointer(uintptr(basePtr) + asu.fields[i].offset)
			*(*unsafe.Pointer)(fieldPtr) = ptr
		}

		// Run update tick
		asu.u.Update(w)
	}
	q.Close()

	// Reset pointer fields to prevent memory leaks
	for i := range asu.fields {
		fieldPtr := unsafe.Pointer(uintptr(basePtr) + asu.fields[i].offset)
		*(*unsafe.Pointer)(fieldPtr) = nil
	}
}

type autoSystemDrawer struct {
	world  *worldImpl
	sys    any
	fields []autoField
	filter arkecs.UnsafeFilter
	d      Drawer
}

func (asd *autoSystemDrawer) Draw(w World, screen *ebiten.Image) {
	basePtr := unsafe.Pointer(reflect.ValueOf(asd.sys).Pointer())

	q := asd.filter.Query()
	for q.Next() {
		// Direct Memory Injection: Write component pointers directly to fields using unsafe offsets.
		for i := range asd.fields {
			ptr := q.Get(asd.fields[i].entry.id)
			fieldPtr := unsafe.Pointer(uintptr(basePtr) + asd.fields[i].offset)
			*(*unsafe.Pointer)(fieldPtr) = ptr
		}

		// Run draw tick
		asd.d.Draw(w, screen)
	}
	q.Close()

	// Reset pointer fields to prevent memory leaks
	for i := range asd.fields {
		fieldPtr := unsafe.Pointer(uintptr(basePtr) + asd.fields[i].offset)
		*(*unsafe.Pointer)(fieldPtr) = nil
	}
}

// ── buildFilterFields ─────────────────────────────────────────────────────────

// buildFilterFields reflects over a system struct and calls build() on any
// field that implements the filterable interface.
func buildFilterFields(system any, ark *arkecs.World) {
	v := reflect.ValueOf(system)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return
	}
	v = v.Elem()
	for i := range v.NumField() {
		structField := v.Type().Field(i)
		if structField.PkgPath != "" {
			continue // skip unexported fields to prevent reflect panic
		}
		field := v.Field(i)
		if !field.CanAddr() {
			continue
		}
		if f, ok := field.Addr().Interface().(filterable); ok {
			f.build(ark)
		}
	}
}

func (w *worldImpl) Spawn(entity any) arkecs.Entity {
	v := reflect.ValueOf(entity)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("engine: Spawn: expected *struct, got %T", entity))
	}
	return w.spawnStruct(v.Elem())
}

func (w *worldImpl) Despawn(e arkecs.Entity) {
	w.ark.RemoveEntity(e)
}

// ── internal ──────────────────────────────────────────────────────────────────

// registerEntry is called from Comp[T].Register.
func (w *worldImpl) registerEntry(
	id arkecs.ID,
	typ reflect.Type,
	copyInto func(unsafe.Pointer, reflect.Value),
	onAdder func(unsafe.Pointer, World, arkecs.Entity),
) {
	if _, ok := w.byType[typ]; ok {
		return
	}
	has := func(e arkecs.Entity) bool {
		return w.ark.Unsafe().Has(e, id)
	}
	get := func(e arkecs.Entity) unsafe.Pointer {
		return w.ark.Unsafe().Get(e, id)
	}
	w.byType[typ] = &compEntry{
		id:       id,
		typ:      typ,
		copyInto: copyInto,
		onAdder:  onAdder,
		has:      has,
		get:      get,
	}
}

// spawnStruct reflects over a struct value, maps its exported fields to
// registered component types, creates the ark entity in one call, copies
// field data into ark storage, then fires OnAdd for each component that
// implements it.
func (w *worldImpl) spawnStruct(sv reflect.Value) arkecs.Entity {
	st := sv.Type()

	type slot struct {
		entry *compEntry
		val   reflect.Value
	}
	var slots []slot
	var ids []arkecs.ID

	for i := range st.NumField() {
		ft := st.Field(i).Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		entry, ok := w.byType[ft]
		if !ok {
			continue
		}
		fv := sv.Field(i)
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				continue
			}
			fv = fv.Elem()
		}
		slots = append(slots, slot{entry, fv})
		ids = append(ids, entry.id)
	}

	if len(ids) == 0 {
		return w.ark.NewEntity()
	}

	// Create the entity with all component IDs at once.
	e := w.ark.Unsafe().NewEntity(ids...)

	for _, s := range slots {
		dst := w.ark.Unsafe().Get(e, s.entry.id)
		if dst == nil {
			continue
		}
		// 1. Copy struct field value into ark storage.
		s.entry.copyInto(dst, s.val)
		// 2. Call OnAdd with fully initialised data.
		if s.entry.onAdder != nil {
			s.entry.onAdder(dst, w, e)
		}
	}

	return e
}

func (w *worldImpl) update() {
	for i := range w.systems {
		if w.systems[i].u != nil {
			w.systems[i].u.Update(w)
		}
	}
}

func (w *worldImpl) draw(screen *ebiten.Image) {
	for i := range w.systems {
		if w.systems[i].d != nil {
			w.systems[i].d.Draw(w, screen)
		}
	}
}
