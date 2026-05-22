package engine

import arkecs "github.com/mlange-42/ark/ecs"

// filterable is implemented by all engine FilterN / MapN types.
// AddSystems detects these fields via reflection and calls build automatically.
type filterable interface {
	build(w *arkecs.World)
}

// Filter1 wraps arkecs.Filter1[A] and is built automatically by AddSystems.
//
//	type MySystem struct {
//	    filter engine.Filter1[Position]
//	}
//	func (s *MySystem) Update(w engine.World) {
//	    q := s.filter.Query()
//	    for q.Next() { pos := q.Get(); ... }
//	}
type Filter1[A any] struct{ inner *arkecs.Filter1[A] }

func (f *Filter1[A]) build(w *arkecs.World)       { f.inner = arkecs.NewFilter1[A](w) }
func (f *Filter1[A]) Query() arkecs.Query1[A]      { return f.inner.Query() }

// Filter2 wraps arkecs.Filter2[A,B].
type Filter2[A, B any] struct{ inner *arkecs.Filter2[A, B] }

func (f *Filter2[A, B]) build(w *arkecs.World)       { f.inner = arkecs.NewFilter2[A, B](w) }
func (f *Filter2[A, B]) Query() arkecs.Query2[A, B]   { return f.inner.Query() }

// Filter3 wraps arkecs.Filter3[A,B,C].
type Filter3[A, B, C any] struct{ inner *arkecs.Filter3[A, B, C] }

func (f *Filter3[A, B, C]) build(w *arkecs.World)          { f.inner = arkecs.NewFilter3[A, B, C](w) }
func (f *Filter3[A, B, C]) Query() arkecs.Query3[A, B, C]   { return f.inner.Query() }

// Filter4 wraps arkecs.Filter4[A,B,C,D].
type Filter4[A, B, C, D any] struct{ inner *arkecs.Filter4[A, B, C, D] }

func (f *Filter4[A, B, C, D]) build(w *arkecs.World)               { f.inner = arkecs.NewFilter4[A, B, C, D](w) }
func (f *Filter4[A, B, C, D]) Query() arkecs.Query4[A, B, C, D]    { return f.inner.Query() }

// Filter5 wraps arkecs.Filter5[A,B,C,D,E].
type Filter5[A, B, C, D, E any] struct{ inner *arkecs.Filter5[A, B, C, D, E] }

func (f *Filter5[A, B, C, D, E]) build(w *arkecs.World)                  { f.inner = arkecs.NewFilter5[A, B, C, D, E](w) }
func (f *Filter5[A, B, C, D, E]) Query() arkecs.Query5[A, B, C, D, E]   { return f.inner.Query() }

