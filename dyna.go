package dyna

// Coalgebra is a nondeterministic F-coalgebra S → List(F S): every way a
// state can shrink into smaller states. A state with no way to shrink has no
// solutions (its value is Zero); a base case is a node without children.
type Coalgebra[S, L any] func(S) []Node[L, S]

// Stats reports how much work Dyna did.
type Stats struct {
	States       int // distinct subproblems solved
	Alternatives int // nodes folded: the edges of the subproblem DAG
}

// Dyna is the dynamorphism: the hylomorphism
//
//	rec = ⊕ ∘ map(alg ∘ F rec) ∘ coalg
//
// memoised on the state, so each subproblem is solved exactly once.
// The coalgebra must be well-founded (no state may reach itself).
func Dyna[S comparable, L, T any](a Algebra[L, T], coalg Coalgebra[S, L], goal S) (T, Stats) {
	memo := map[S]T{}
	var st Stats
	var rec func(S) T
	rec = func(s S) T {
		if v, ok := memo[s]; ok {
			return v
		}
		acc := a.Zero
		for _, node := range coalg(s) {
			st.Alternatives++
			acc = a.Plus(acc, a.Alg(Fmap(rec, node)))
		}
		memo[s] = acc
		st.States++
		return acc
	}
	return rec(goal), st
}

// Tree is the initial algebra μF: one concrete solution.
type Tree[L any] struct{ Node Node[L, *Tree[L]] }

// Cata folds one tree with an algebra (the catamorphism).
func Cata[L, T any](alg func(Node[L, T]) T, t *Tree[L]) T {
	return alg(Fmap(func(c *Tree[L]) T { return Cata(alg, c) }, t.Node))
}

// Unfold grows every tree the coalgebra allows from s, without sharing.
// It is exponential: use it to test, not to solve.
func Unfold[S, L any](coalg Coalgebra[S, L], s S) []*Tree[L] {
	var out []*Tree[L]
	for _, n := range coalg(s) {
		if len(n.Kids) == 0 {
			out = append(out, &Tree[L]{Node[L, *Tree[L]]{Label: n.Label}})
			continue
		}
		// cartesian product of the children's trees
		combos := [][]*Tree[L]{{}}
		for _, kid := range n.Kids {
			sub := Unfold(coalg, kid)
			var next [][]*Tree[L]
			for _, c := range combos {
				for _, t := range sub {
					next = append(next, append(append([]*Tree[L](nil), c...), t))
				}
			}
			combos = next
		}
		for _, c := range combos {
			out = append(out, &Tree[L]{Node[L, *Tree[L]]{Label: n.Label, Kids: c}})
		}
	}
	return out
}

// BruteForce is the specification Dyna must meet: fold every tree on its
// own, then ⊕ the results ("the sum over all paths of the product along each").
func BruteForce[S, L, T any](a Algebra[L, T], coalg Coalgebra[S, L], s S) (T, int) {
	trees := Unfold(coalg, s)
	acc := a.Zero
	for _, t := range trees {
		acc = a.Plus(acc, Cata(a.Alg, t))
	}
	return acc, len(trees)
}
