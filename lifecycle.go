package engine

import arkecs "github.com/mlange-42/ark/ecs"

// OnAdder may be implemented by a component to run initialization logic
// immediately after it is placed on an entity and its data is written.
//
//	func (c *Collider) OnAdd(w engine.World, e arkecs.Entity) {
//	    engine.GetResource[*resolv.Space](w).Add(c.Shape)
//	}
type OnAdder interface {
	OnAdd(w World, e arkecs.Entity)
}

// OnRemover may be implemented by a component to run teardown logic just
// before the entity (or this component) is removed from the world.
//
//	func (c *Collider) OnRemove(w engine.World, e arkecs.Entity) {
//	    if c.Shape.Space() != nil { c.Shape.Space().Remove(c.Shape) }
//	}
type OnRemover interface {
	OnRemove(w World, e arkecs.Entity)
}
