package bimax

import (
	"fmt"

	"github.com/yourbasic/graph"
)

// BiMaxBinaryMatrix takes in an n by m binary matrix where n is the number of
// rows and m the number of columns.  The data for the binary matrix is
// specified as a slice of uint8 containing only 1's and 0's.
func BiMaxBinaryMatrix(n, m int, data []uint8) *BiMaxResult {
	if (len(data) / m) != n {
		panic(fmt.Sprintf("matrix data cannot be reshaped into [%d, %d]", n, m))
	}
	G := graph.New(n + m)
	U, V := NewSet(), NewSet()

	for i, x := range data {
		switch x {
		case 0:
			continue
		case 1:
			// Calculate graph index.
			graphIdxRow := i / m
			graphIdxCol := i%m + n
			// Add to graph and each bipartite vertex set.
			U.Add(graphIdxRow)
			V.Add(graphIdxCol)
			G.AddBoth(graphIdxRow, graphIdxCol)
		default:
			panic(fmt.Sprintf("%d is not a zero or 1", x))
		}
	}
	// Get the result of the bimax Function
	return BiMax(G, U, V)
}

// BiMaxVertices takes in two slices of verticies uu and vv that represents sets
// U and V in a bipartite graph G such that a vertex within one of these sets
// does not form an edge any other vertex within the same set.
// U = {u∈U | (u,uʹ)∉G}
// V = {v∈V | (v,vʹ)∉G}
// The edge set of U and V makes up the graph G such that every vertex in set U
// must map to some vertex in set V and vice versa
func BiMaxVertices(uu, vv []int) *BiMaxResult {
	if len(uu) != len(vv) {
		panic(fmt.Sprintf("len(uu): %d len(vv): %d must be equal", len(uu), len(vv)))
	}
	U, V := NewSetFromSlice(uu), NewSetFromSlice(vv)
	// Find the maximal vertex index in the set ov V to be used the max number of
	// vertecies in the Graph
	var vtxCount int
	V.Each(func(v int) (_ bool) {
		if v > vtxCount {
			vtxCount = v
		}
		return
	})
	G := graph.New(vtxCount + 1)
	for i := 0; i < len(uu); i++ {
		G.AddBoth(uu[i], vv[i])
	}
	return BiMax(G, U, V)
}

// BiMaxResult represents the result returned from 'BiMax' as a set of boths
// rows and columns or as the set of two vertecies in a the maximal biclique of
// graph G.
type BiMaxResult struct {
	Rows, Cols *SetOp
}

// BiMax finds the maximal bipartitie clique of a bipartite graph of graph G
// where G is a bipartite graph of (U ∪ V, E(G))
func BiMax(G *graph.Mutable, L, PU *UnorderedSet) *BiMaxResult {
	// L: is a set of verticies ∈ U that are common neigbors of R; initially L = U
	// R: is a set of verticies ∈ V belonging to the current biclique; initially
	// empty
	R := NewSet()
	// P: is a set of verticies ∈ V that can be added to R, initially P = V,
	// sorted by non-decreasing order of neigborhood size
	P := PU.Order(func(v1, v2 int) bool {
		return G.Degree(v2) <= G.Degree(v1)
	})

	// Q: is a set of verticies used to determine maximality, initially empty
	Q := NewSet()

	// Resulting sets
	Rows, Cols := NewSet(), NewSet()

	var bicliqueFind func(P *OrderedSet, L, R, Q *UnorderedSet)
	bicliqueFind = func(P *OrderedSet, L, R, Q *UnorderedSet) {
		for i := 0; i < P.Card(); i++ {
			x := P.Get(0)

			// Candidates
			c := NewSetWith(x)
			// Rʹ is set of verticies in current biclique
			Rʹ := R.Union(c)
			// Lʹ is the set verticies in L that neighbor x
			Lʹ := NeighborSet(x, L.SetOp, G, false).(*UnorderedSet)
			// Complement of Lʹ
			Lʹᶜ := L.Difference(Lʹ)

			// Create new sets for P and Q
			Pʹ, Qʹ := P.New().(*OrderedSet), Q.New().(*UnorderedSet)

			// var bicliqueFind func(Pʹ *OrderedSet, Lʹ, Rʹ, Qʹ *UnorderedSet)
			// bicliqueFind = func(Pʹ *OrderedSet, Lʹ, Rʹ, Qʹ *UnorderedSet) {
			maximal := true
			// For all v in Q
			Q.Each(func(v int) (done bool) {
				// Cardinality of closed neighborhood at v is the the degree + 1
				LʹNeighborVDegree := NeighborSetDegree(v, Lʹ.SetOp, G, false)
				// N := NeighborSet(*v, Lʹ, G, false)
				if LʹNeighborVDegree == Lʹ.Card() {
					maximal = false
					return true
				}
				if LʹNeighborVDegree > 0 {
					Qʹ.Add(v)
				}
				return
			})

			if maximal {
				// For each v in P excluding x
				P.Each(func(v int) (done bool) {
					if x == v {
						return
					}

					// Set of {uϵLʹ| (u, v) ϵ E(G)} set of verticies u such that u and v
					// are edges in graph G including v
					// N := NeighborSet(*v, Lʹ, G, false)
					N := NeighborSet(v, Lʹ.SetOp, G, false)
					if N.Card() == Lʹ.Card() {
						Rʹ.Add(v)
						// Set of {uϵLʹᶜ| (u, v) ϵ E(G)} set of verticies u such that u and v
						// are edges in graph G
						if NeighborSetDegree(v, Lʹᶜ.SetOp, G, false) == 0 {
							c.Add(v)
						}
						return
					}
					if N.Card() == 0 {
						Pʹ.Add(v)
					}
					return
				})
				// TODO: might be able to optimize based on number of enumerated
				// bicliques <20-01-21, Max Schulte> //

				// Print Maximal biclique
				// fmt.Println(Lʹ, Rʹ)

				if (Rows.Card() * Cols.Card()) < (Lʹ.Card() * Rʹ.Card()) {
					// A = Lʹ.Copy()
					// B = Rʹ.Copy()
					Rows = Lʹ
					Cols = Rʹ
				}
				// fmt.Println(Lʹ, Rʹ)
				if Pʹ.Card() == 0 {
					bicliqueFind(Pʹ, Lʹ, Rʹ, Qʹ)
				}
			}
			CValues := c.Values()
			Q.Update(CValues...)
			P.Remove(CValues...)
		}
	}
	bicliqueFind(P, L, R, Q)
	result := BiMaxResult{&SetOp{Rows}, &SetOp{Cols}}
	return &result
}

// ClosedDegree returns the degree of the closed neighborhood at v
func ClosedDegree(v int, G *graph.Mutable) (degree int) {
	degree = G.Degree(v)
	if degree > 0 {
		degree++
	}
	return
}

func NeighborSetDegree(v int, set *SetOp, G *graph.Mutable, closed bool) int {
	result := 0
	set.Each(func(u int) (done bool) {
		if !G.Edge(u, v) {
			return
		}
		result++
		return
	})
	if (result > 0) && closed {
		result++
	}
	return result
}

func NeighborSet(v int, set *SetOp, G *graph.Mutable, closed bool) Set {
	result := &SetOp{set.New()}
	if closed {
		result.Add(v)
	}
	set.Each(func(u int) (done bool) {
		if !G.Edge(u, v) {
			return
		}
		result.Add(u)
		return
	})
	return result.Set
}
