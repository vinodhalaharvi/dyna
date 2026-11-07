package dyna_test

import (
	"fmt"

	"github.com/vinodhalaharvi/dyna"
)

type coin struct {
	Done bool
	Coin int
}

func coins(a int) []dyna.Node[coin, int] {
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

func Example() {
	ways, _ := dyna.Dyna(dyna.Ungraded(dyna.Count[coin]()), coins, 11)

	fewest := dyna.TopK(1, func(l coin, kids []float64, _ float64, _ []float64) float64 {
		if l.Done {
			return 0
		}
		return kids[0] - 1
	})
	best, _ := dyna.Dyna(dyna.Ungraded(fewest), coins, 11)
	var used []int
	for e := best[0]; !e.Label.Done; e = e.Kids[0] {
		used = append(used, e.Label.Coin)
	}
	fmt.Println(ways, -best[0].Score, used)
	// Output: 218 3 [1 5 5]
}
