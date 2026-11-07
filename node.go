// Package dyna writes dynamic programs as recursion schemes over semirings.
//
// A problem is described by three things:
//
//	a functor     Node[L, X]   one step: a label and its children
//	a coalgebra   S → []Node[L, S]   every way a state shrinks (the "projections")
//	an algebra    Node[L, T] → T, over a commutative monoid (T, ⊕)   how answers combine
//
// and Dyna folds the coalgebra's unfolding with the algebra, collapsing
// alternatives with ⊕ and memoising on the state. That is dynamic programming.
// It is correct exactly when the algebra distributes over ⊕, which CheckLaws
// tests and BruteForce confirms.
package dyna

// Node is one step of a polynomial functor F X = Σ_label X^arity(label):
// a label and its children. Go has no higher-kinded types, so every functor
// is encoded this way; the label says which summand, Kids holds the X's.
//
// A collage split is Node{Label: "split H", Kids: [left, right]};
// a list step is Node{Label: "keep 7", Kids: [rest]}; a leaf has no Kids.
type Node[L, X any] struct {
	Label L
	Kids  []X
}

// Fmap is the functor's action on arrows: map every child, keep the label.
func Fmap[L, X, Y any](f func(X) Y, n Node[L, X]) Node[L, Y] {
	if len(n.Kids) == 0 {
		return Node[L, Y]{Label: n.Label}
	}
	kids := make([]Y, len(n.Kids))
	for i, k := range n.Kids {
		kids[i] = f(k)
	}
	return Node[L, Y]{Label: n.Label, Kids: kids}
}
