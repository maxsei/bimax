package main

import (
	"fmt"

	"github.com/yourbasic/graph"
)

// Pick vertex v
// Create a set of its neighbors u
// Create an empty set for neighbors'
func IntSort(v1, v2 *int) bool {
	return *v1 <= *v2
}

var depth int

// BiMax enumeration of graph G where:
// G: is a bipartite graph of (U ∪ V, E)
func BiMax(G *graph.Mutable) (Set, Set) {
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
	PUnordered := E.Difference(L)
	if L.Card() < PUnordered.Card() {
		var tmp *UnorderedSet
		tmp = L
		L = PUnordered
		PUnordered = tmp
	}
	// R: is a set of verticies ∈ V belonging to the current biclique; initially
	// empty
	R := NewSet()
	// P: is a set of verticies ∈ V that can be added to R, initially P = V,
	// sorted by non-decreasing order of neigborhood size
	P := PUnordered.Order(func(v1, v2 *int) bool {
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

	var bicliqueFind func(P *OrderedSet, L, R, Q *UnorderedSet)
	bicliqueFind = func(P *OrderedSet, L, R, Q *UnorderedSet) {
		// PValues := P.Values()

		for i := 0; 0 < P.Card(); i++ {
			x, _ := P.Iterator().Iter()
			// x := P.Iterator().Get(i)
			// x := P.Pop()
			// for i, x := range PValues {
			fmt.Printf("\n|%-10sP.Card(): %d depth: %-10d|\n", "", P.Card(), depth)
			// x := *P.Pop()

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
				fmt.Println(Lʹ, Rʹ)
				if Pʹ.Card() == 0 {
					// fmt.Printf("R = %+v\n", R)
					// fmt.Printf("Q = %+v\n", Q)
					// fmt.Printf("P = %+v\n", P)
					// fmt.Printf("L = %+v\n", L)
					// fmt.Printf("E = %+v\n", E)
					// fmt.Printf("Rʹ = %+v\n", Rʹ)
					// fmt.Printf("Qʹ = %+v\n", Qʹ)
					// fmt.Printf("Pʹ = %+v\n", Pʹ)
					// fmt.Printf("Lʹᶜ = %+v\n", Lʹᶜ)
					// fmt.Printf("Lʹ = %+v\n", Lʹ)
					// fmt.Printf("C = %+v\n", C)
					depth++
					bicliqueFind(Pʹ, Lʹ, Rʹ, Qʹ)
					depth--
				}
			}
			CValues := C.Values()
			fmt.Printf("CValues = %+v\n", CValues)
			Q.Update(CValues...)
			P.Remove(CValues...)
		}
	}
	bicliqueFind(P, L, R, Q)
	return nil, nil
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
