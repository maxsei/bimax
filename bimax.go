package main

import (
	"fmt"

	"github.com/yourbasic/graph"
)

// BiMax enumeration of graph G where:
// G: is a bipartite graph of (U ∪ V, E)
func BiMax(G graph.Iterator) (R, P *Set) {
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
	// sorted by increasing order of neigborhood size
	P = L.Difference(E)

	// // Sort P by neighbors
	// for iterator := P.Iterator(); ; {
	// 	v, done := iterator.Iter()
	// 	if done {
	// 		break
	// 	}
	// 	n := Neighbors(v, G)
	// }

	// Q: is a set of verticies used to determine maximality, initially empty
	Q := NewSet()

	fmt.Printf("G = %+v\n", G)
	fmt.Printf("E = %+v\n", E)
	fmt.Printf("L = %+v\n", L)
	fmt.Printf("R = %+v\n", R)
	fmt.Printf("P = %+v\n", P)
	fmt.Printf("Q = %+v\n", Q)

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

func Neighbors(v int, g graph.Iterator) *Set {
	result := NewSet()
	g.Visit(v, func(u int, _ int64) bool {
		result.Add(u)
		return false
	})
	return result
}

func biMax(G graph.Iterator, L, R, P, Q *Set) bool {
	return true
}
