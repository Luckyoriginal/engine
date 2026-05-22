package engine

import (
	"reflect"
	"unsafe"

	arkecs "github.com/mlange-42/ark/ecs"
)

// Comp is a typed component descriptor for type T.
//
// Declare one package-level variable per component type, call Register once
// in Setup, then use the ark API directly for filters and mappers:
//
//	var PosComp = engine.NewComp[Position]()
//
//	func (g *Game) Setup(w engine.World) {
//	    PosComp.Register(w)
//	    filter := arkecs.NewFilter2[Position, Velocity](w.Ark())
//	}
type Comp[T any] struct {
	id    arkecs.ID
	valid bool
}

// NewComp allocates an unregistered Comp descriptor.
func NewComp[T any]() *Comp[T] { return &Comp[T]{} }

// Register wires T into the world. Safe to call more than once.
//
// What happens here:
//  1. ark assigns a stable component ID for T.
//  2. A copyInto func is stored so Spawn can copy struct fields into ark storage.
//  3. If *T implements OnAdder, an onAdder callback is stored and called by
//     Spawn after data is written (so OnAdd always sees real values).
//  4. If *T implements OnRemover, an ark OnRemoveEntity observer is registered
//     so Despawn triggers OnRemove automatically.
func (c *Comp[T]) Register(w World) {
	if c.valid {
		return
	}
	impl := w.(*worldImpl)
	c.id = arkecs.ComponentID[T](&impl.ark)
	c.valid = true

	typ := reflect.TypeOf((*T)(nil)).Elem()

	copyInto := func(dst unsafe.Pointer, src reflect.Value) {
		*(*T)(dst) = src.Interface().(T)
	}

	var onAdder func(unsafe.Pointer, World, arkecs.Entity)

	if reflect.PointerTo(typ).Implements(reflect.TypeOf((*OnAdder)(nil)).Elem()) {
		onAdder = func(ptr unsafe.Pointer, w World, e arkecs.Entity) {
			reflect.NewAt(typ, ptr).Interface().(OnAdder).OnAdd(w, e)
		}
	}

	if reflect.PointerTo(typ).Implements(reflect.TypeOf((*OnRemover)(nil)).Elem()) {
		// OnRemover is wired as an ark observer so it fires on Despawn too,
		// not just on Spawn-paired removes.
		// .With(arkecs.C[T]()) filters the observer to entities that carry T.
		id := c.id
		arkecs.Observe(arkecs.OnRemoveEntity).
			With(arkecs.C[T]()).
			Do(func(entity arkecs.Entity) {
				ptr := impl.ark.Unsafe().Get(entity, id)
				if ptr == nil {
					return
				}
				reflect.NewAt(typ, ptr).Interface().(OnRemover).OnRemove(impl, entity)
			}).
			Register(&impl.ark)
	}

	impl.registerEntry(c.id, typ, copyInto, onAdder)
}

// ID returns the ark component ID. Panics if called before Register.
func (c *Comp[T]) ID() arkecs.ID {
	if !c.valid {
		panic("engine: Comp.ID() called before Register")
	}
	return c.id
}
