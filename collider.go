package engine

import (
	arkecs "github.com/mlange-42/ark/ecs"
	"github.com/solarlune/resolv"
)

// Collider is a first-class component that wraps a resolv.IShape.
// It automatically registers and unregisters its shape with the world's *resolv.Space
// resource using the engine's built-in OnAdd and OnRemove lifecycle hooks.
type Collider struct {
	Shape resolv.IShape
}

// OnAdd is called automatically when the Collider is spawned onto an entity.
func (c *Collider) OnAdd(w World, e arkecs.Entity) {
	space := GetResource[*resolv.Space](w)
	space.Add(c.Shape)
}

// OnRemove is called automatically when the Collider is removed or despawned.
func (c *Collider) OnRemove(w World, e arkecs.Entity) {
	space := GetResource[*resolv.Space](w)
	space.Remove(c.Shape)
}

// OnIntersect runs an intersection test against shapes near this collider (within a 1-cell margin).
func (c *Collider) OnIntersect(callback func(set resolv.IntersectionSet) bool) {
	c.Shape.IntersectionTest(resolv.IntersectionTestSettings{
		TestAgainst: c.Shape.SelectTouchingCells(1).FilterShapes(),
		OnIntersect: callback,
	})
}

// OnIntersectWithTag runs an intersection test against nearby shapes matching the specified tag.
func (c *Collider) OnIntersectWithTag(tag resolv.Tags, callback func(set resolv.IntersectionSet) bool) {
	c.Shape.IntersectionTest(resolv.IntersectionTestSettings{
		TestAgainst: c.Shape.SelectTouchingCells(1).FilterShapes().ByTags(tag),
		OnIntersect: callback,
	})
}
