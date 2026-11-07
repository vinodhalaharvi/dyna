package dyna

const maxArity = 4

// Entry is one solution kept by TopK: its score and the tree that earned it,
// so the winning structure can be read back out (a witness).
type Entry[L any] struct {
	Score float64
	Label L
	Kids  []*Entry[L]
}

// Ranked is a TopK value: at most K entries, best first.
type Ranked[L any] []*Entry[L]

// Score combines a node's label, its children's scores and grades into the
// node's score. It must be monotone (non-decreasing) in every child score:
// that is what makes TopK distribute over ⊕.
type Score[L any] func(label L, kids []float64, self float64, grades []float64) float64

// A TopK algebra reuses internal buffers: use one value from one goroutine.
//
// TopK keeps the k best solutions by score: the (max, ⊗) semiring lifted to
// k-best lists, with witnesses. ⊕ merges two lists; a node scores every
// combination of its children's entries and keeps the k best.
func TopK[L any](k int, score Score[L]) GradedAlgebra[L, Ranked[L]] {
	plus := func(a, b Ranked[L]) Ranked[L] {
		if len(b) == 0 {
			return a
		}
		if len(a) == 0 {
			return b
		}
		if len(a) == k && b[0].Score <= a[k-1].Score {
			return a // nothing in b makes the cut; ties keep a, as the merge would
		}
		out := make(Ranked[L], 0, min(k, len(a)+len(b)))
		i, j := 0, 0
		for len(out) < k && (i < len(a) || j < len(b)) {
			if j >= len(b) || (i < len(a) && a[i].Score >= b[j].Score) {
				out = append(out, a[i])
				i++
			} else {
				out = append(out, b[j])
				j++
			}
		}
		return out
	}
	type cand struct {
		s   float64
		idx [maxArity]int
	}
	var scores []float64
	var cands []cand
	// slabs: entries and pointer lists are carved from big blocks, so a
	// solve makes thousands of allocations instead of millions
	var entSlab []Entry[L]
	var ptrSlab []*Entry[L]
	ptrs := func(n int) []*Entry[L] {
		if len(ptrSlab) < n {
			ptrSlab = make([]*Entry[L], max(n, 8192))
		}
		p := ptrSlab[:n:n]
		ptrSlab = ptrSlab[n:]
		return p
	}
	alg := func(n Node[L, Ranked[L]], self float64, grades []float64) Ranked[L] {
		if len(n.Kids) == 0 {
			return Ranked[L]{{Score: score(n.Label, nil, self, nil), Label: n.Label}}
		}
		for _, kid := range n.Kids {
			if len(kid) == 0 {
				return nil
			}
		}
		if len(n.Kids) > maxArity {
			panic("dyna.TopK: nodes may have at most 4 children")
		}
		// enumerate the product of the children's lists (each at most k long)
		cands = cands[:0]
		var idx [maxArity]int
		for {
			scores = scores[:0]
			for p, kid := range n.Kids {
				scores = append(scores, kid[idx[p]].Score)
			}
			cands = append(cands, cand{score(n.Label, scores, self, grades), idx})
			p := len(n.Kids) - 1
			for p >= 0 {
				idx[p]++
				if idx[p] < len(n.Kids[p]) {
					break
				}
				idx[p] = 0
				p--
			}
			if p < 0 {
				break
			}
		}
		// stable selection of the k best (ties keep enumeration order)
		m := min(k, len(cands))
		for i := 0; i < m; i++ {
			best := i
			for j := i + 1; j < len(cands); j++ {
				if cands[j].s > cands[best].s {
					best = j
				}
			}
			if best != i {
				c := cands[best]
				copy(cands[i+1:best+1], cands[i:best])
				cands[i] = c
			}
		}
		out := Ranked[L](ptrs(m))
		if len(entSlab) < m {
			entSlab = make([]Entry[L], 4096)
		}
		ents := entSlab[:m:m]
		entSlab = entSlab[m:]
		for c := 0; c < m; c++ {
			kids := ptrs(len(n.Kids))
			for p := range n.Kids {
				kids[p] = n.Kids[p][cands[c].idx[p]]
			}
			ents[c] = Entry[L]{Score: cands[c].s, Label: n.Label, Kids: kids}
			out[c] = &ents[c]
		}
		return out
	}
	return GradedAlgebra[L, Ranked[L]]{
		Monoid: Monoid[Ranked[L]]{Zero: nil, Plus: plus, IsZero: func(r Ranked[L]) bool { return len(r) == 0 }},
		Alg:    alg,
	}
}
