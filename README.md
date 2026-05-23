# engine

A small ECS layer for [ebiten](https://ebitengine.org/) games, backed by [ark](https://github.com/mlange-42/ark)'s high-performance archetype engine.

Components, systems, and entities are declared as plain Go structs.  
Filters are built automatically. Lifecycle hooks are optional.

---

## Installation

```
go get github.com/Luckyoriginal/engine
```

---

## Quick start

```go
package main

import (
    "image/color"
    "log"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/Luckyoriginal/engine"
)

// 1. Define components as plain structs.
type Pos struct{ X, Y float32 }
type Vel struct{ X, Y float32 }

// 2. Declare one descriptor per component.
var (
    PosComp = engine.NewComp[Pos]()
    VelComp = engine.NewComp[Vel]()
)

// 3. Bundle components into an entity struct.
type Ball struct {
    Pos
    Vel
}

// 4. Write systems. Declare engine.FilterN fields — they are built
//    automatically when the system is registered.
type Move struct{ filter engine.Filter2[Pos, Vel] }

func (s *Move) Update(_ engine.World) {
    q := s.filter.Query()
    for q.Next() {
        pos, vel := q.Get()
        pos.X += vel.X
        pos.Y += vel.Y
        if pos.X < 0 || pos.X > 640 { vel.X = -vel.X }
        if pos.Y < 0 || pos.Y > 480 { vel.Y = -vel.Y }
    }
}

type Draw struct{ filter engine.Filter1[Pos] }

func (s *Draw) Draw(_ engine.World, screen *ebiten.Image) {
    screen.Fill(color.RGBA{20, 20, 40, 255})
    q := s.filter.Query()
    for q.Next() {
        pos := q.Get()
        vector.FillCircle(screen, pos.X, pos.Y, 12, color.White, true)
    }
}

// 5. Implement engine.Scene.
type Game struct{}

func (g *Game) Setup(w engine.World) {
    PosComp.Register(w)
    VelComp.Register(w)

    w.AddSystems(&Move{}, &Draw{})

    w.Spawn(&Ball{
        Pos: Pos{X: 320, Y: 240},
        Vel: Vel{X: 3, Y: 2},
    })
}

func main() {
    ebiten.SetWindowSize(640, 480)
    if err := ebiten.RunGame(engine.NewGame(&Game{})); err != nil {
        log.Fatal(err)
    }
}
```

---

## Concepts

### Components

Any plain Go struct. No embedding, no interface required.

```go
type Health struct{ Current, Max int }
type Sprite struct{ Image *ebiten.Image }
```

Register each one with a descriptor before spawning entities that carry it:

```go
var HealthComp = engine.NewComp[Health]()

func (g *Game) Setup(w engine.World) {
    HealthComp.Register(w)
}
```

---

### Entities

A struct whose exported fields are registered component types. Pass a pointer to `w.Spawn`.

```go
type Player struct {
    Pos
    Vel
    Health
}

w.Spawn(&Player{
    Pos:    Pos{X: 100, Y: 100},
    Health: Health{Current: 100, Max: 100},
})
```

Fields that are not registered components are silently ignored, so you can include helper data in the struct without issue.

`Spawn` returns an `arkecs.Entity` handle you can store to despawn or query later:

```go
e := w.Spawn(&Player{...})
// later:
w.Despawn(e)
```

---

### Systems

A struct that implements `Update`, `Draw`, or both.

```go
// Update-only system
type Physics struct{ filter engine.Filter2[Pos, Vel] }

func (s *Physics) Update(_ engine.World) { ... }

// Draw-only system
type Renderer struct{ filter engine.Filter1[Pos] }

func (s *Renderer) Draw(_ engine.World, screen *ebiten.Image) { ... }

// Both
type HUD struct{}

func (s *HUD) Update(_ engine.World) { ... }
func (s *HUD) Draw(_ engine.World, screen *ebiten.Image) { ... }
```

Register with `w.AddSystems`. Systems run in the order they are added, update systems before draw systems.

```go
w.AddSystems(&Physics{}, &Renderer{}, &HUD{})
```

---

### Filters

Declare `engine.FilterN` fields on your system struct. They are initialised automatically by `AddSystems` — no `Init` method needed.

```go
type MySystem struct {
    filter engine.Filter3[Pos, Vel, Health]
}

func (s *MySystem) Update(_ engine.World) {
    q := s.filter.Query()
    for q.Next() {
        pos, vel, hp := q.Get()
        _ = pos; _ = vel; _ = hp
    }
}
```

Filters `Filter1` through `Filter5` are provided. If you need more components, use `w.Ark()` to access the underlying ark world directly.

---

### Zero-Boilerplate Systems (Auto-Entity Systems)

For maximum simplicity and zero boilerplate, you can design systems that operate directly on individual entities. Instead of declaring an explicit `FilterN` field and writing a manual `for q.Next() { ... }` loop, you simply declare exported pointer fields of the component types you need.

The engine automatically detects these fields at setup time, constructs an optimized archetype filter under the hood, and iterates through matching entities on every tick:

```go
type Movement struct {
    Position *Pos  
    Velocity *Vel  
}

// Called automatically for each entity that carries BOTH Pos and Vel.
func (s *Movement) Update(w engine.World) {
    s.Position.X += s.Velocity.X
    s.Position.Y += s.Velocity.Y
}
```

#### Performance Characteristics
* **0% Reflection per Tick**: Component field detection and unsafe memory offsets are cached once at startup. During the tick, raw pointers are written directly using `unsafe.Pointer` writes, resulting in zero GC allocation pressure.
* **Archetype Speed**: Leverages `arkecs` native archetype queries. The update loop skips all non-matching entities entirely, delivering true $O(M)$ native compiled ECS execution speeds.

---

### Lifecycle hooks — `OnAdd` and `OnRemove`

Implement these optional methods on a component's pointer receiver to run code when the component is attached to or removed from an entity.

A common use-case is a physics collider that needs to register/unregister with an external space:

```go
type Collider struct {
    Shape *resolv.Circle
}

func (c *Collider) OnAdd(w engine.World, e arkecs.Entity) {
    engine.GetResource[*resolv.Space](w).Add(c.Shape)
}

func (c *Collider) OnRemove(w engine.World, e arkecs.Entity) {
    if c.Shape.Space() != nil {
        c.Shape.Space().Remove(c.Shape)
    }
}
```

`OnAdd` is called immediately after `Spawn` writes the component data, so `c.Shape` is already populated.  
`OnRemove` is called just before the entity is destroyed by `Despawn`.

Both are optional. Components that don't need them are unaffected.

---

### Resources

World-level singletons — physics spaces, asset stores, timers, score, anything shared across systems.

```go
// store (before any OnAdd that reads it)
engine.SetResource(w, resolv.NewSpace(640, 480, 16, 16))

// retrieve (anywhere you have a World)
space := engine.GetResource[*resolv.Space](w)
```

---

### Manual system init — `Init`

If a system needs to do something beyond filter setup at registration time (e.g. fetch a resource and store it), implement `Init`:

```go
type EnemyAI struct {
    filter engine.Filter2[Pos, EnemyTag]
    space  *resolv.Space
}

func (s *EnemyAI) Init(w engine.World) {
    s.space = engine.GetResource[*resolv.Space](w)
}
```

---

### Scene Manager (Type-Safe State & Transitions)

For larger games, managing transitions between a Main Menu, Gameplay, and a Game Over screen is critical. The engine includes a type-safe **Scene Manager** backed by [stagehand](https://github.com/joelschutz/stagehand) that allows you to easily switch between scenes, each with its own isolated ECS `World` (so entities and systems are completely unloaded when switching).

Scenes share a type-safe State struct that is passed from scene to scene during transitions.

#### 1. Define your shared game state
```go
type GameState struct {
    Score int
}
```

#### 2. Implement the `engine.ECSScene[T]` interface
Each scene implements `Setup` (where you get a fresh ECS `World` and the current state) and `Unload` (where you return the updated state to pass to the next scene).

```go
type MenuScene struct {
    state GameState
}

func (s *MenuScene) Setup(w engine.World, state GameState, sm *engine.SceneManager[GameState]) {
    s.state = state
    w.AddSystems(&MenuSystem{sm: sm, state: &s.state})
}

func (s *MenuScene) Unload() GameState {
    return s.state
}
```

#### 3. Transition with effects
You can transition immediately using `sm.SwitchTo(nextScene)` or with visual transitions (like **Fade** or **Slide**) using `sm.SwitchWithTransition(nextScene, transition)`:

```go
import "github.com/joelschutz/stagehand"

// Inside a System:
if spacePressed {
    // Transition to Gameplay with a Fade effect (5% alpha per frame)
    sys.sm.SwitchWithTransition(&GameplayScene{}, stagehand.NewFadeTransition[GameState](0.05))
}
```

#### 4. Run the Scene Manager
Instantiate the manager with the initial scene and state, then run it with Ebiten:

```go
func main() {
    ebiten.SetWindowSize(640, 480)
    sm := engine.NewSceneManager[GameState](&MenuScene{}, GameState{Score: 0})
    if err := ebiten.RunGame(sm); err != nil {
        log.Fatal(err)
    }
}
```

---

### Direct ark access

For anything beyond `Filter5`, or for advanced queries (exclusion filters, optional components), use `w.Ark()` to get the underlying `*arkecs.World` and use ark's API directly.

```go
import arkecs "github.com/mlange-42/ark/ecs"

filter := arkecs.NewFilter6[A, B, C, D, E, F](w.Ark())
```

---

## File overview

| File | Contents |
|---|---|
| `comp.go` | `Comp[T]`, `NewComp[T]()`, `RegisterComp[T]()` |
| `filter.go` | `Filter1` … `Filter5` |
| `lifecycle.go` | `OnAdder`, `OnRemover` |
| `system.go` | `Updater`, `Drawer`, `Initer` |
| `resource.go` | `SetResource[T]`, `GetResource[T]` |
| `world.go` | `World` interface + implementation |
| `game.go` | `Scene`, `Game`, `NewGame` (single-scene) |
| `scene.go` | `ECSScene[T]`, `SceneManager[T]` (multi-scene) |
