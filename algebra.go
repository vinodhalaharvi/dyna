package dyna

// Monoid is the "choose" half of a semiring: ⊕ with identity Zero.
// Laws: associative, commutative, Zero is the identity.
type Monoid[T any] struct {
	Zero   T
	Plus   func(a, b T) T
	IsZero func(T) bool
}

// Algebra is an F-algebra Node[L, T] → T over a commutative monoid (T, ⊕).
// Its law is distributivity, which for a functor means multilinearity:
// in every child position k,
//
//	Alg(… x⊕y …) = Alg(… x …) ⊕ Alg(… y …)    and    Alg(… 0 …) = 0.
type Algebra[L, T any] struct {
	Monoid[T]
	Alg func(Node[L, T]) T
}

// GradedAlgebra is an algebra whose step may read each child's grade value
// (an aspect ratio, a duration). Lift turns it into an ordinary Algebra over
// grade-indexed vectors. For a node without children, grades is empty and the
// node's own grade value is passed in self.
//
// Alg must not keep n.Kids or grades: Lift reuses those slices.
type GradedAlgebra[L, T any] struct {
	Monoid[T]
	Alg func(n Node[L, T], self float64, grades []float64) T
}

// Pair is the carrier of a product algebra.
type Pair[A, B any] struct {
	A A
	B B
}

// Product runs two graded algebras in one pass, componentwise.
// It obeys the laws whenever both factors do.
func Product[L, A, B any](x GradedAlgebra[L, A], y GradedAlgebra[L, B]) GradedAlgebra[L, Pair[A, B]] {
	var ka []A
	var kb []B
	return GradedAlgebra[L, Pair[A, B]]{
		Monoid: Monoid[Pair[A, B]]{
			Zero:   Pair[A, B]{x.Zero, y.Zero},
			Plus:   func(p, q Pair[A, B]) Pair[A, B] { return Pair[A, B]{x.Plus(p.A, q.A), y.Plus(p.B, q.B)} },
			IsZero: func(p Pair[A, B]) bool { return x.IsZero(p.A) && y.IsZero(p.B) },
		},
		Alg: func(n Node[L, Pair[A, B]], self float64, grades []float64) Pair[A, B] {
			ka, kb = ka[:0], kb[:0]
			for _, k := range n.Kids {
				ka = append(ka, k.A)
				kb = append(kb, k.B)
			}
			na := Node[L, A]{Label: n.Label, Kids: ka}
			nb := Node[L, B]{Label: n.Label, Kids: kb}
			return Pair[A, B]{x.Alg(na, self, grades), y.Alg(nb, self, grades)}
		},
	}
}

// Count is (+, ×): the number of solutions. A node counts the product of its
// children's counts (1 for a node without children).
func Count[L any]() GradedAlgebra[L, float64] {
	return GradedAlgebra[L, float64]{
		Monoid: Monoid[float64]{Zero: 0, Plus: func(a, b float64) float64 { return a + b }, IsZero: func(x float64) bool { return x == 0 }},
		Alg: func(n Node[L, float64], _ float64, _ []float64) float64 {
			p := 1.0
			for _, k := range n.Kids {
				p *= k
			}
			return p
		},
	}
}

// Ungraded uses a graded algebra where no grading is needed: every grade
// value it sees is 0.
func Ungraded[L, T any](g GradedAlgebra[L, T]) Algebra[L, T] {
	var zeros []float64
	return Algebra[L, T]{Monoid: g.Monoid, Alg: func(n Node[L, T]) T {
		for len(zeros) < len(n.Kids) {
			zeros = append(zeros, 0)
		}
		return g.Alg(n, 0, zeros[:len(n.Kids)])
	}}
}
