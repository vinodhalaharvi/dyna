package dyna

import (
	"math/rand"
	"testing"
)

// Coin change: a state is an amount; it shrinks by using one coin.
type coin struct {
	Done bool
	Coin int
}

func coins(cs []int) Coalgebra[int, coin] {
	return func(a int) []Node[coin, int] {
		if a == 0 {
			return []Node[coin, int]{{Label: coin{Done: true}}}
		}
		var out []Node[coin, int]
		for _, c := range cs {
			if a-c >= 0 {
				out = append(out, Node[coin, int]{Label: coin{Coin: c}, Kids: []int{a - c}})
			}
		}
		return out
	}
}

func TestCoinChangeCount(t *testing.T) {
	alg := Ungraded(Count[coin]())
	got, st := Dyna(alg, coins([]int{1, 2, 5}), 11)
	want, trees := BruteForce(alg, coins([]int{1, 2, 5}), 11)
	if got != 218 || want != 218 {
		t.Fatalf("got %v and %v, want 218", got, want)
	}
	t.Logf("218 ordered ways: %d trees by brute force, %d states and %d steps by Dyna", trees, st.States, st.Alternatives)
}

// 0/1 knapsack graded by weight: state i decides item i.
type pick struct {
	End  bool
	Item int
	Keep bool
}

func TestKnapsackMatchesBruteForce(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for trial := 0; trial < 30; trial++ {
		n := 1 + r.Intn(10)
		w := make([]int, n)
		v := make([]float64, n)
		for i := range w {
			w[i], v[i] = 1+r.Intn(6), r.Float64()
		}
		coalg := func(i int) []Node[pick, int] {
			if i == n {
				return []Node[pick, int]{{Label: pick{End: true}}}
			}
			return []Node[pick, int]{
				{Label: pick{Item: i, Keep: true}, Kids: []int{i + 1}},
				{Label: pick{Item: i}, Kids: []int{i + 1}},
			}
		}
		grading := Grading[pick]{N: 16, Value: func(g int) float64 { return float64(g) },
			Of: func(l pick, kids []int) int {
				if l.End {
					return 0
				}
				g := kids[0]
				if l.Keep {
					g += w[l.Item]
				}
				if g >= 16 {
					return -1
				}
				return g
			}}
		score := func(l pick, kids []float64, _ float64, _ []float64) float64 {
			if l.End {
				return 0
			}
			if l.Keep {
				return kids[0] + v[l.Item]
			}
			return kids[0]
		}
		alg := Lift(Product(Count[pick](), TopK(3, score)), grading)
		dp, _ := Dyna(alg, coalg, 0)
		bf, _ := BruteForce(alg, coalg, 0)
		for g := 0; g < 16; g++ {
			a, b := dp.Get(g, alg.Zero.Get(0, Pair[float64, Ranked[pick]]{})), bf.Get(g, Pair[float64, Ranked[pick]]{})
			if a.A != b.A || !RankedScoresEqual(a.B, b.B) {
				t.Fatalf("trial %d weight %d: dyna %v/%d, brute force %v/%d", trial, g, a.A, len(a.B), b.A, len(b.B))
			}
		}
	}
}

func TestLaws(t *testing.T) {
	r := rand.New(rand.NewSource(2))
	labels := []pick{{Item: 0, Keep: true}, {Item: 1}}
	arity := func(pick) int { return 1 }
	score := func(l pick, kids []float64, _ float64, _ []float64) float64 { return kids[0] + 0.5 }
	var res []LawResult
	res = append(res, CheckLaws(Count[pick](), LawSpec[pick, float64]{Name: "Count", Labels: labels, Arity: arity,
		Gen: func(r *rand.Rand) float64 { return float64(r.Intn(6)) }, Eq: func(a, b float64) bool { return a == b }}, 5000, r)...)
	res = append(res, CheckLaws(TopK(3, score), LawSpec[pick, Ranked[pick]]{Name: "TopK", Labels: labels, Arity: arity,
		Gen: GenRanked(3, pick{End: true}), Eq: RankedScoresEqual[pick]}, 5000, r)...)
	// a binary node: bottleneck score, as in collage
	type cut struct{ H bool }
	bott := func(l cut, kids []float64, _ float64, g []float64) float64 {
		f := g[0] / (g[0] + g[1])
		return min(kids[0]*f, kids[1]*(1-f))
	}
	res = append(res, CheckLaws(TopK(3, bott), LawSpec[cut, Ranked[cut]]{Name: "TopK bottleneck", Labels: []cut{{true}},
		Arity: func(cut) int { return 2 }, Gen: GenRanked(3, cut{}), Eq: RankedScoresEqual[cut]}, 5000, r)...)
	for _, l := range res {
		if !l.OK {
			t.Errorf("%s: %s fails", l.Algebra, l.Law)
		}
	}
}

// A non-monotone score must be caught.
func TestLawsCatchBadAlgebra(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	bad := func(l pick, kids []float64, _ float64, _ []float64) float64 { return -kids[0] }
	res := CheckLaws(TopK(2, bad), LawSpec[pick, Ranked[pick]]{Name: "bad", Labels: []pick{{}}, Arity: func(pick) int { return 1 },
		Gen: GenRanked(3, pick{End: true}), Eq: RankedScoresEqual[pick]}, 5000, r)
	for _, l := range res {
		if l.Law == "distributes over ⊕" && l.OK {
			t.Fatal("a score that reverses order was not caught")
		}
	}
}
