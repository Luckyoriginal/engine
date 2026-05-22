package engine

import "github.com/hajimehoshi/ebiten/v2"

// Scene is implemented by the top-level game struct.
// Setup is called once before the first Update tick.
type Scene interface {
	Setup(w World)
}

// Game wraps worldImpl and implements ebiten.Game.
//
//	g := engine.NewGame(&MyScene{})
//	ebiten.RunGame(g)
type Game struct {
	w     *worldImpl
	scene Scene
	ready bool
}

// NewGame creates a Game from a Scene implementation.
func NewGame(scene Scene) *Game {
	return &Game{w: newWorldImpl(), scene: scene}
}

// World returns the World for pre-game setup (e.g. storing resources before
// the first tick).
func (g *Game) World() World { return g.w }

// Update implements ebiten.Game.
func (g *Game) Update() error {
	if !g.ready {
		g.scene.Setup(g.w)
		g.ready = true
	}
	g.w.update()
	return nil
}

// Draw implements ebiten.Game.
func (g *Game) Draw(screen *ebiten.Image) {
	g.w.draw(screen)
}

// Layout implements ebiten.Game.
func (g *Game) Layout(ow, oh int) (int, int) { return ow, oh }
