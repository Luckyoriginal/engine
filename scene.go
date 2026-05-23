package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joelschutz/stagehand"
)

// ECSScene is implemented by scenes that manage their own ECS world.
type ECSScene[T any] interface {
	// Setup is called when the scene is loaded.
	// You can register components, spawn entities, add systems, and store resources.
	Setup(w World, state T, sm *SceneManager[T])

	// Unload is called when switching away from this scene.
	// It should return the state to be passed to the next scene.
	Unload() T
}

// SceneManager manages multiple ECSScenes with a shared state.
// It implements ebiten.Game and wraps stagehand.SceneManager under the hood.
type SceneManager[T any] struct {
	inner *stagehand.SceneManager[T]
}

// NewSceneManager creates a new SceneManager starting with the given initial scene and state.
func NewSceneManager[T any](initialScene ECSScene[T], state T) *SceneManager[T] {
	sm := &SceneManager[T]{}
	wrapper := &ecsSceneWrapper[T]{
		world: newWorldImpl(),
		scene: initialScene,
		sm:    sm,
	}
	sm.inner = stagehand.NewSceneManager[T](wrapper, state)
	return sm
}

// SwitchTo changes the current scene immediately.
func (sm *SceneManager[T]) SwitchTo(nextScene ECSScene[T]) {
	wrapper := &ecsSceneWrapper[T]{
		world: newWorldImpl(),
		scene: nextScene,
		sm:    sm,
	}
	sm.inner.SwitchTo(wrapper)
}

// SwitchWithTransition changes the current scene using the specified transition.
func (sm *SceneManager[T]) SwitchWithTransition(nextScene ECSScene[T], transition stagehand.SceneTransition[T]) {
	wrapper := &ecsSceneWrapper[T]{
		world: newWorldImpl(),
		scene: nextScene,
		sm:    sm,
	}
	sm.inner.SwitchWithTransition(wrapper, transition)
}

// Update implements ebiten.Game.
func (sm *SceneManager[T]) Update() error {
	return sm.inner.Update()
}

// Draw implements ebiten.Game.
func (sm *SceneManager[T]) Draw(screen *ebiten.Image) {
	sm.inner.Draw(screen)
}

// Layout implements ebiten.Game.
func (sm *SceneManager[T]) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return sm.inner.Layout(outsideWidth, outsideHeight)
}

// ── ecsSceneWrapper ───────────────────────────────────────────────────────────

// ecsSceneWrapper wraps an ECSScene to implement stagehand.Scene.
type ecsSceneWrapper[T any] struct {
	world *worldImpl
	scene ECSScene[T]
	sm    *SceneManager[T]
	ready bool
}

func (w *ecsSceneWrapper[T]) Update() error {
	w.world.update()
	return nil
}

func (w *ecsSceneWrapper[T]) Draw(screen *ebiten.Image) {
	w.world.draw(screen)
}

func (w *ecsSceneWrapper[T]) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (w *ecsSceneWrapper[T]) Load(state T, _ stagehand.SceneController[T]) {
	if !w.ready {
		w.scene.Setup(w.world, state, w.sm)
		w.ready = true
	}
}

func (w *ecsSceneWrapper[T]) Unload() T {
	return w.scene.Unload()
}
