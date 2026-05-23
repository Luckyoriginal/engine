package engine

import arkecs "github.com/mlange-42/ark/ecs"

// filterable is implemented by all engine FilterN / MapN types.
// AddSystems detects these fields via reflection and calls build automatically.
type filterable interface {
	build(w *arkecs.World)
}

// Filter1 wraps arkecs.Filter1[A] and is built automatically by AddSystems.
type Filter1[A any] struct{ inner *arkecs.Filter1[A] }

func (f *Filter1[A]) build(w *arkecs.World)       { f.inner = arkecs.NewFilter1[A](w) }
func (f *Filter1[A]) Query() arkecs.Query1[A]      { return f.inner.Query() }

// Each runs the given callback on every matching entity in the query.
func (f *Filter1[A]) Each(forEach func(a *A)) {
	q := f.Query()
	for q.Next() {
		forEach(q.Get())
	}
}

// Filter2 wraps arkecs.Filter2[A,B].
type Filter2[A, B any] struct{ inner *arkecs.Filter2[A, B] }

func (f *Filter2[A, B]) build(w *arkecs.World)       { f.inner = arkecs.NewFilter2[A, B](w) }
func (f *Filter2[A, B]) Query() arkecs.Query2[A, B]   { return f.inner.Query() }

// Each runs the given callback on every matching entity in the query.
func (f *Filter2[A, B]) Each(forEach func(a *A, b *B)) {
	q := f.Query()
	for q.Next() {
		a, b := q.Get()
		forEach(a, b)
	}
}

// Filter3 wraps arkecs.Filter3[A,B,C].
type Filter3[A, B, C any] struct{ inner *arkecs.Filter3[A, B, C] }

func (f *Filter3[A, B, C]) build(w *arkecs.World)          { f.inner = arkecs.NewFilter3[A, B, C](w) }
func (f *Filter3[A, B, C]) Query() arkecs.Query3[A, B, C]   { return f.inner.Query() }

// Each runs the given callback on every matching entity in the query.
func (f *Filter3[A, B, C]) Each(forEach func(a *A, b *B, c *C)) {
	q := f.Query()
	for q.Next() {
		a, b, c := q.Get()
		forEach(a, b, c)
	}
}

// Filter4 wraps arkecs.Filter4[A,B,C,D].
type Filter4[A, B, C, D any] struct{ inner *arkecs.Filter4[A, B, C, D] }

func (f *Filter4[A, B, C, D]) build(w *arkecs.World)               { f.inner = arkecs.NewFilter4[A, B, C, D](w) }
func (f *Filter4[A, B, C, D]) Query() arkecs.Query4[A, B, C, D]    { return f.inner.Query() }

// Each runs the given callback on every matching entity in the query.
func (f *Filter4[A, B, C, D]) Each(forEach func(a *A, b *B, c *C, d *D)) {
	q := f.Query()
	for q.Next() {
		a, b, c, d := q.Get()
		forEach(a, b, c, d)
	}
}

// Filter5 wraps arkecs.Filter5[A,B,C,D,E].
type Filter5[A, B, C, D, E any] struct{ inner *arkecs.Filter5[A, B, C, D, E] }

func (f *Filter5[A, B, C, D, E]) build(w *arkecs.World)                  { f.inner = arkecs.NewFilter5[A, B, C, D, E](w) }
func (f *Filter5[A, B, C, D, E]) Query() arkecs.Query5[A, B, C, D, E]   { return f.inner.Query() }

// Each runs the given callback on every matching entity in the query.
func (f *Filter5[A, B, C, D, E]) Each(forEach func(a *A, b *B, c *C, d *D, e *E)) {
	q := f.Query()
	for q.Next() {
		a, b, c, d, e := q.Get()
		forEach(a, b, c, d, e)
	}
}
