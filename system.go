package engine

import "github.com/hajimehoshi/ebiten/v2"

// Updater is implemented by systems that run logic every tick.
type Updater interface {
	Update(w World)
}

// Drawer is implemented by systems that render every frame.
// A system may implement both Updater and Drawer.
type Drawer interface {
	Draw(w World, screen *ebiten.Image)
}

// Initer may be implemented by a system to receive the World once,
// called automatically by AddSystems before the first Update/Draw.
// Use it to build filters and mappers so the system struct stays clean:
//
//	type Movement struct {
//	    filter arkecs.Filter2[Position, Velocity]
//	}
//	func (s *Movement) Init(w engine.World) {
//	    s.filter = arkecs.NewFilter2[Position, Velocity](w.Ark())
//	}
//	func (s *Movement) Update(w engine.World) { ... }
type Initer interface {
	Init(w World)
}

type system struct {
	u Updater
	d Drawer
}

func wrapSystem(s any) system {
	u, _ := s.(Updater)
	d, _ := s.(Drawer)
	if u == nil && d == nil {
		panic("engine: AddSystems: value implements neither Updater nor Drawer")
	}
	return system{u, d}
}
