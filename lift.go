package dyna

// Grading indexes answers by a grade: an aspect-ratio bucket, a number of
// seconds. Grades are 0..N-1 and combine through Of, which is the action of a
// monoid on grades (aspect ratios add side by side, durations add in a list).
type Grading[L any] struct {
	N     int                           // number of grades
	Value func(grade int) float64       // what the algebra sees for a grade
	Of    func(label L, kids []int) int // a node's grade from its children's; -1 drops it
}

// Vec is an element of the monoid semiring T[G]: a finitely supported map
// from grades to T, like a polynomial whose exponents are grades.
type Vec[T any] struct {
	At []T   // dense coefficients; nil is the zero vector
	NZ []int // support, in insertion order
}

// Get returns the coefficient at grade g (the algebra's Zero if absent).
func (v Vec[T]) Get(g int, zero T) T {
	if v.At == nil || g < 0 || g >= len(v.At) {
		return zero
	}
	return v.At[g]
}

// Lift turns a graded algebra into an algebra over grade-indexed vectors.
// A node's vector is the convolution of its children's: every combination of
// child grades lands at the grade Of computes, and coefficients there are ⊕-ed.
// Lift preserves the laws: if g distributes over ⊕, so does the lifted algebra.
func Lift[L, T any](g GradedAlgebra[L, T], gr Grading[L]) Algebra[L, Vec[T]] {
	add := func(v *Vec[T], b int, x T) {
		if g.IsZero(x) {
			return
		}
		if v.At == nil {
			v.At = make([]T, gr.N)
			for i := range v.At {
				v.At[i] = g.Zero
			}
		}
		if g.IsZero(v.At[b]) {
			v.NZ = append(v.NZ, b)
			v.At[b] = x
			return
		}
		v.At[b] = g.Plus(v.At[b], x)
	}
	plus := func(a, b Vec[T]) Vec[T] {
		if a.At == nil {
			return b
		}
		if b.At == nil {
			return a
		}
		out := Vec[T]{At: append([]T(nil), a.At...), NZ: append([]int(nil), a.NZ...)}
		for _, i := range b.NZ {
			add(&out, i, b.At[i])
		}
		return out
	}
	var (
		grades []int
		values []float64
		coeffs []T
	)
	alg := func(n Node[L, Vec[T]]) Vec[T] {
		var out Vec[T]
		k := len(n.Kids)
		if k == 0 {
			if b := gr.Of(n.Label, nil); b >= 0 {
				add(&out, b, g.Alg(Node[L, T]{Label: n.Label}, gr.Value(b), nil))
			}
			return out
		}
		for _, kid := range n.Kids {
			if len(kid.NZ) == 0 {
				return out
			}
		}
		grades, values, coeffs = grades[:0], values[:0], coeffs[:0]
		for range n.Kids {
			grades = append(grades, 0)
			values = append(values, 0)
			coeffs = append(coeffs, g.Zero)
		}
		var pos [maxArity]int // odometer over the children's supports
		if k > maxArity {
			panic("dyna.Lift: nodes may have at most 4 children")
		}
		for {
			for p, kid := range n.Kids {
				b := kid.NZ[pos[p]]
				grades[p], values[p], coeffs[p] = b, gr.Value(b), kid.At[b]
			}
			if b := gr.Of(n.Label, grades); b >= 0 {
				add(&out, b, g.Alg(Node[L, T]{Label: n.Label, Kids: coeffs}, gr.Value(b), values))
			}
			p := k - 1
			for p >= 0 {
				pos[p]++
				if pos[p] < len(n.Kids[p].NZ) {
					break
				}
				pos[p] = 0
				p--
			}
			if p < 0 {
				return out
			}
		}
	}
	return Algebra[L, Vec[T]]{
		Monoid: Monoid[Vec[T]]{Zero: Vec[T]{}, Plus: plus, IsZero: func(v Vec[T]) bool { return len(v.NZ) == 0 }},
		Alg:    alg,
	}
}
