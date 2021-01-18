package main

import (
	"fmt"

	"github.com/yourbasic/graph"
)

// BiMax enumeration of graph G where:
// G: is a bipartite graph of (U ∪ V, E)
func BiMax(G *graph.Mutable) (R, P Set) {
	// Travers
	// E: is the set of all verticies in G ( Edge set )
	E := NewSet()
	parts, _ := graph.Bipartition(G)
	for _, v := range parts {
		E.Add(v)
		G.Visit(v, func(w int, c int64) (skip bool) {
			E.Add(w)
			return
		})
	}

	// L: is a set of verticies ∈ U that are common neigbors of R; initially L = U
	L := NewSetFromSlice(parts)
	// R: is a set of verticies ∈ V belonging to the current biclique; initially
	// empty
	R = NewSet()
	// P: is a set of verticies ∈ V that can be added to R, initially P = V,
	// sorted by non-decreasing order of neigborhood size
	P = (E.Difference(L)).Order(func(v1, v2 *int) bool {
		return G.Degree(*v1) <= G.Degree(*v2)
	})

	// Q: is a set of verticies used to determine maximality, initially empty
	Q := NewSet()

	// fmt.Printf("G = %+v\n", G)
	fmt.Printf("E = %+v\n", E)
	fmt.Printf("L = %+v\n", L)
	fmt.Printf("R = %+v\n", R)
	fmt.Printf("P = %+v\n", P)
	fmt.Printf("Q = %+v\n", Q)

	for P.Card() < 0 {

	}

	// TODO: sort values in p by neighborhood size <15-01-21, Max Schulte> //
	// for P.Card() != 0{
	// }

	// pIterator := P.Iterator()
	// for {
	// 	x, done := pIterator.Iter()
	// 	if done {
	// 		break
	// 	}
	// 	RPrime := R.Intersection(NewSetWith(x))
	// 	LPrime := NewSet()

	// 	PPrime, QPrime := NewSet(), NewSet()
	// }
	return
}
