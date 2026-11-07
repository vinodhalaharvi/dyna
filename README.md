# dyna

Dynamic programming in Go, written as recursion schemes over semirings.

```sh
go get github.com/vinodhalaharvi/dyna
```

Every dynamic program is the same three pieces:

| Piece | In dyna | What it answers |
|---|---|---|
| a **shape** | `Node[L, X]`: a label and its children | what one step of a solution looks like |
| a **coalgebra** | `Coalgebra[S, L]`: `S → []Node[L, S]` | how a problem shrinks into smaller ones |
| an **algebra** | `Algebra[L, T]`: `Node[L, T] → T` over a monoid `(T, ⊕)` | how answers combine, and which one you want |

`Dyna` folds the unfolding of the coalgebra with the algebra, collapses alternatives with `⊕`, and memoises on the state:

```
rec = ⊕ ∘ map(alg ∘ F rec) ∘ coalg        (a dynamorphism)
```

Change the algebra and the same code answers a different question. Count solutions with `Count` (+, ×), keep the best k with `TopK`, do both at once with `Product`, and index every answer by a grade such as a length or an aspect ratio with `Lift`.

## Example: coin change, two questions, one coalgebra

```go
type coin struct{ Done bool; Coin int }

coins := func(a int) []dyna.Node[coin, int] {
	if a == 0 {
		return []dyna.Node[coin, int]{{Label: coin{Done: true}}}
	}
	var out []dyna.Node[coin, int]
	for _, c := range []int{1, 2, 5} {
		if a-c >= 0 {
			out = append(out, dyna.Node[coin, int]{Label: coin{Coin: c}, Kids: []int{a - c}})
		}
	}
	return out
}

ways, _ := dyna.Dyna(dyna.Ungraded(dyna.Count[coin]()), coins, 11) // 218

fewest := dyna.TopK(1, func(l coin, kids []float64, _ float64, _ []float64) float64 {
	if l.Done {
		return 0
	}
	return kids[0] - 1 // each coin costs one; TopK maximises
})
best, _ := dyna.Dyna(dyna.Ungraded(fewest), coins, 11) // score -3: three coins, with the tree to prove it
```

## The law that makes it correct

Dynamic programming keeps only the best answer to each subproblem. That is safe exactly when the algebra **distributes over ⊕**:

```
a ⊗ (b ⊕ c) = (a ⊗ b) ⊕ (a ⊗ c)
```

This is optimal substructure, stated as an equation. For a functor it means every node is linear in each child. `CheckLaws` property-tests it on random values, along with the monoid laws and annihilation by zero, and `BruteForce` gives a reference to compare against: it unfolds every solution separately (exponential, so for tests only) and must equal `Dyna`.

A score that is not monotone, like `max` with multiplication by negative numbers, breaks the law. The test suite includes such an algebra and checks that `CheckLaws` catches it.

## Pieces

| | |
|---|---|
| `Node`, `Fmap` | the polynomial functor `F X = Σ label · X^arity`; Go has no higher-kinded types, so every shape is a label plus children |
| `Coalgebra`, `Dyna`, `Stats` | the dynamorphism and how much work it did (states, edges) |
| `Tree`, `Cata`, `Unfold`, `BruteForce` | the specification: every solution, folded one by one |
| `Monoid`, `Algebra`, `GradedAlgebra`, `Ungraded` | the semiring interface |
| `Count`, `TopK`, `Product` | (+, ×); k-best with witnesses; two algebras in one pass |
| `Grading`, `Lift`, `Vec` | the monoid semiring `T[G]`: answers indexed by a grade that combines through a monoid |
| `CheckLaws`, `GenRanked`, `RankedScoresEqual` | property tests for your algebra |

## Used by

- [collage](https://github.com/vinodhalaharvi/collage): guillotine photo layouts, graded by aspect ratio
- [trim](https://github.com/vinodhalaharvi/trim): cut a talk to an exact length, graded by duration

## License

MIT
