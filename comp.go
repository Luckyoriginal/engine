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

// RegisterComp registers type T as a component in the world.
func RegisterComp[T any](w World) {
	impl := w.(*worldImpl)
	typ := reflect.TypeOf((*T)(nil)).Elem()

	if _, ok := impl.byType[typ]; ok {
		return
	}

	id := arkecs.ComponentID[T](&impl.ark)

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

	impl.registerEntry(id, typ, copyInto, onAdder)
}

// Register wires T into the world. Safe to call more than once.
func (c *Comp[T]) Register(w World) {
	if c.valid {
		return
	}
	RegisterComp[T](w)
	impl := w.(*worldImpl)
	typ := reflect.TypeOf((*T)(nil)).Elem()
	if entry, ok := impl.byType[typ]; ok {
		c.id = entry.id
		c.valid = true
	}
}

// ID returns the ark component ID. Panics if called before Register.
func (c *Comp[T]) ID() arkecs.ID {
	if !c.valid {
		panic("engine: Comp.ID() called before Register")
	}
	return c.id
}
