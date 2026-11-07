package dyna

import (
	"math"
	"math/rand"
)

// LawResult is the outcome of one law on one algebra.
type LawResult struct {
	Algebra string `json:"algebra"`
	Law     string `json:"law"`
	OK      bool   `json:"ok"`
	Trials  int    `json:"trials"`
}

// LawSpec says how to test a graded algebra: labels with children to try,
// how to make random elements, and when two elements are equal.
type LawSpec[L, T any] struct {
	Name   string
	Labels []L // labels of nodes with children
	Arity  func(L) int
	Gen    func(*rand.Rand) T
	Eq     func(a, b T) bool
	Grade  func(*rand.Rand) float64 // random grade values; default (0.2, 5.2)
}

// CheckLaws property-tests the laws Dyna relies on: ⊕ is a commutative
// monoid, and every node is multilinear in its children (distributes over ⊕
// in each position) and annihilated by Zero.
func CheckLaws[L, T any](g GradedAlgebra[L, T], spec LawSpec[L, T], trials int, r *rand.Rand) []LawResult {
	grade := spec.Grade
	if grade == nil {
		grade = func(r *rand.Rand) float64 { return 0.2 + 5*r.Float64() }
	}
	// step applies the node to kids with kid p replaced by x
	step := func(l L, kids []T, p int, x T, self float64, gs []float64) T {
		ks := append([]T(nil), kids...)
		ks[p] = x
		return g.Alg(Node[L, T]{Label: l, Kids: ks}, self, gs)
	}
	type law struct {
		name string
		ok   func() bool
	}
	pick := func() (L, []T, int, float64, []float64) {
		l := spec.Labels[r.Intn(len(spec.Labels))]
		n := spec.Arity(l)
		kids := make([]T, n)
		gs := make([]float64, n)
		for i := range kids {
			kids[i], gs[i] = spec.Gen(r), grade(r)
		}
		return l, kids, r.Intn(n), grade(r), gs
	}
	laws := []law{
		{"⊕ associative", func() bool {
			x, y, z := spec.Gen(r), spec.Gen(r), spec.Gen(r)
			return spec.Eq(g.Plus(g.Plus(x, y), z), g.Plus(x, g.Plus(y, z)))
		}},
		{"⊕ commutative", func() bool { x, y := spec.Gen(r), spec.Gen(r); return spec.Eq(g.Plus(x, y), g.Plus(y, x)) }},
		{"0 is ⊕-identity", func() bool { x := spec.Gen(r); return spec.Eq(g.Plus(x, g.Zero), x) }},
		{"distributes over ⊕", func() bool {
			l, kids, p, self, gs := pick()
			x, y := spec.Gen(r), spec.Gen(r)
			return spec.Eq(step(l, kids, p, g.Plus(x, y), self, gs),
				g.Plus(step(l, kids, p, x, self, gs), step(l, kids, p, y, self, gs)))
		}},
		{"0 annihilates", func() bool {
			l, kids, p, self, gs := pick()
			return g.IsZero(step(l, kids, p, g.Zero, self, gs))
		}},
	}
	var out []LawResult
	for _, lw := range laws {
		ok := true
		for i := 0; i < trials && ok; i++ {
			ok = lw.ok()
		}
		out = append(out, LawResult{spec.Name, lw.name, ok, trials})
	}
	return out
}

// RankedScoresEqual compares two TopK values by their scores (ties may pick
// different witnesses, which is fine).
func RankedScoresEqual[L any](a, b Ranked[L]) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(a[i].Score-b[i].Score) > 1e-9 {
			return false
		}
	}
	return true
}

// GenRanked makes a random TopK value of up to k leaf entries, with ties.
func GenRanked[L any](k int, label L) func(*rand.Rand) Ranked[L] {
	return func(r *rand.Rand) Ranked[L] {
		n := r.Intn(k + 1)
		out := make(Ranked[L], n)
		s := 1.0
		for i := range out {
			s *= 0.3 + 0.7*r.Float64()
			if i > 0 && r.Intn(4) == 0 {
				s = out[i-1].Score
			}
			out[i] = &Entry[L]{Score: s, Label: label}
		}
		return out
	}
}
