package main

import (
	"fmt"

	"github.com/yourbasic/graph"
)

func BiMaxBinaryMatrix(n, m int, data []uint8) (rows, cols []int) {
	if (len(data) / m) != n {
		panic(fmt.Sprintf("matrix data cannot be reshaped into [%d, %d]", n, m))
	}
	G := graph.New(n + m)
	uu, vv := NewSet(), NewSet()
	for i, x := range data {
		if x == 0 {
			continue
		}
		// Calculate graph index.
		graphIdxRow := i / m
		graphIdxCol := i%m + n
		// Add to graph and each bipartite vertex set.
		uu.Add(graphIdxRow)
		vv.Add(graphIdxCol)
		G.AddBoth(graphIdxRow, graphIdxCol)
	}
	// Get the result of the bimax Function
	A, B := BiMax(G, uu, vv)
	// Find which contains rows
	rowset, colset := A, B
	if B.Has(0) {
		rowset, colset = B, A
	}
	rows = rowset.Values()
	cols = colset.Values()
	for i := 0; i < len(cols); i++ {
		cols[i] -= n
	}
	return
}

// func BiMaxVerticies(uu, vv []int) (Set, Set) {
// 	if len(uu) != len(vv) {
// 		panic(fmt.Sprintf("expected uu and vv to be the same length got len(uu): %d len(vv) %d", len(uu), len(vv)))
// 	}
// 	G := graph.New(len(uu))
// 	for i := 0; i < len(uu); i++ {
// 		G.AddBoth(uu[i], vv[i])
// 	}
// 	A, B := NewSetFromSlice(uu), NewSetFromSlice(vv)
// 	if A.Card() < B.Card() {
// 		return BiMax(B, A, PU)
// 	}
// 	return BiMax(A, B, PU)
// }

// BiMax finds the maximal bipartitie clique of a bipartite graph of graph G
// where G is a bipartite graph of (U ∪ V, E)
func BiMax(G *graph.Mutable, L, PU *UnorderedSet) (*SetOp, *SetOp) {
	// L: is a set of verticies ∈ U that are common neigbors of R; initially L = U
	// R: is a set of verticies ∈ V belonging to the current biclique; initially
	// empty
	R := NewSet()
	// P: is a set of verticies ∈ V that can be added to R, initially P = V,
	// sorted by non-decreasing order of neigborhood size
	P := PU.Order(func(v1, v2 *int) bool {
		return G.Degree(*v1) <= G.Degree(*v2)
	})

	// Q: is a set of verticies used to determine maximality, initially empty
	Q := NewSet()

	// Resulting sets
	A, B := NewSet(), NewSet()

	var bicliqueFind func(P *OrderedSet, L, R, Q *UnorderedSet)
	bicliqueFind = func(P *OrderedSet, L, R, Q *UnorderedSet) {
		// PValues := P.Values()

		for i := 0; 0 < P.Card(); i++ {
			x, _ := P.Iterator().Iter()

			// Candidates
			C := NewSetWith(*x)
			// Rʹ is set of verticies in current biclique
			Rʹ := R.Union(C)
			// Lʹ is the set verticies in L that neighbor x
			Lʹ := NeighborSet(*x, L, G, false).(*UnorderedSet)
			// Complement of Lʹ
			Lʹᶜ := L.Difference(Lʹ)

			// Create new sets for P and Q
			Pʹ, Qʹ := P.New().(*OrderedSet), Q.New().(*UnorderedSet)

			// var bicliqueFind func(Pʹ *OrderedSet, Lʹ, Rʹ, Qʹ *UnorderedSet)
			// bicliqueFind = func(Pʹ *OrderedSet, Lʹ, Rʹ, Qʹ *UnorderedSet) {
			maximal := true
			// For all v in Q
			for iterator := Q.Iterator(); ; {
				v, done := iterator.Iter()
				if done {
					break
				}
				// Cardinality of closed neighborhood at v is the the degree + 1
				LʹNeighborVDegree := NeighborSetDegree(*v, Lʹ, G, false)
				// N := NeighborSet(*v, Lʹ, G, false)
				if LʹNeighborVDegree == Lʹ.Card() {
					maximal = false
					break
				}
				if LʹNeighborVDegree > 0 {
					Qʹ.Add(*v)
				}
			}
			if maximal {
				// For each v in P excluding x
				for iterator := P.Iterator(); ; {
					v, done := iterator.Iter()
					if done {
						break
					}
					if *x == *v {
						continue
					}

					// Set of {uϵLʹ| (u, v) ϵ E(G)} set of verticies u such that u and v
					// are edges in graph G including v
					// N := NeighborSet(*v, Lʹ, G, false)
					N := NeighborSet(*v, Lʹ, G, false)
					if N.Card() == Lʹ.Card() {
						Rʹ.Add(*v)
						// Set of {uϵLʹᶜ| (u, v) ϵ E(G)} set of verticies u such that u and v
						// are edges in graph G
						if NeighborSetDegree(*v, Lʹᶜ, G, false) == 0 {
							C.Add(*v)
						}
						continue
					}
					if N.Card() == 0 {
						Pʹ.Add(*v)
					}
				}
				// Print Maximal biclique
				if (A.Card() * B.Card()) < (Lʹ.Card() * Rʹ.Card()) {
					// A = Lʹ.Copy()
					// B = Rʹ.Copy()
					A = Lʹ
					B = Rʹ
				}
				// fmt.Println(Lʹ, Rʹ)
				if Pʹ.Card() == 0 {
					bicliqueFind(Pʹ, Lʹ, Rʹ, Qʹ)
				}
			}
			CValues := C.Values()
			Q.Update(CValues...)
			P.Remove(CValues...)
		}
	}
	bicliqueFind(P, L, R, Q)
	return &SetOp{A}, &SetOp{B}
}

// ClosedDegree returns the degree of the closed neighborhood at v
func ClosedDegree(v int, G *graph.Mutable) (degree int) {
	degree = G.Degree(v)
	if degree > 0 {
		degree++
	}
	return
}

func NeighborSetDegree(v int, set Set, G *graph.Mutable, closed bool) int {
	result := 0
	for iterator := set.Iterator(); ; {
		u, done := iterator.Iter()
		if done {
			break
		}
		if !G.Edge(*u, v) {
			continue
		}
		result++
	}
	if (result > 0) && closed {
		result++
	}
	return result
}

func NeighborSet(v int, set Set, G *graph.Mutable, closed bool) Set {
	result := &SetOp{set.New()}
	if closed {
		result.Add(v)
	}
	for iterator := set.Iterator(); ; {
		u, done := iterator.Iter()
		if done {
			break
		}
		if !G.Edge(*u, v) {
			continue
		}
		result.Add(*u)
	}
	return result.Set
}
