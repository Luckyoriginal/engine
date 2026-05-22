package engine

import "fmt"

// SetResource stores a typed singleton value in the world.
// Use the type itself as the implicit key:
//
//	engine.SetResource(w, myPhysicsSpace)
func SetResource[T any](w World, v T) {
	w.SetResource(resKey[T]{}, v)
}

// GetResource retrieves a typed singleton. Panics if not set.
//
//	space := engine.GetResource[*resolv.Space](w)
func GetResource[T any](w World) T {
	v, ok := w.Resource(resKey[T]{})
	if !ok {
		var zero T
		panic(fmt.Sprintf("engine: resource %T not found; call SetResource first", zero))
	}
	return v.(T)
}

// resKey is a zero-size unique map key per type T.
type resKey[T any] struct{}
