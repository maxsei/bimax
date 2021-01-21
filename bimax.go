package main

import (
	"fmt"

	"C"

	"github.com/yourbasic/graph"
)
import (
	"reflect"
	"unsafe"
)

//export BiMaxBinaryMatrixC
func BiMaxBinaryMatrixC(n64, m64 int64, input *C.char) (C.size_t, *C.longlong, C.size_t, *C.longlong) {
	n := int(n64)
	m := int(m64)
	// Convert C input data into Go data
	var data []uint8
	dataH := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	dataH.Data = uintptr(unsafe.Pointer(input))
	dataH.Len = n * m

	result := BiMaxBinaryMatrix(n, m, data)
	return result.ToC()
}

// BiMaxBinaryMatrix takes in an n by m binary matrix where n is the number of
// rows and m the number of columns.  The data for the binary matrix is
// specified as a slice of uint8 containing only 1's and 0's.
func BiMaxBinaryMatrix(n, m int, data []uint8) *BiMaxResult {
	if (len(data) / m) != n {
		panic(fmt.Sprintf("matrix data cannot be reshaped into [%d, %d]", n, m))
	}
	fmt.Printf("n = %+v\n", n)
	fmt.Printf("m = %+v\n", m)
	G := graph.New(n + m)
	U, V := NewSet(), NewSet()
	for i, x := range data {
		if x == 0 {
			continue
		}
		// Calculate graph index.
		graphIdxRow := i / m
		graphIdxCol := i%m + n
		// Add to graph and each bipartite vertex set.
		U.Add(graphIdxRow)
		V.Add(graphIdxCol)
		G.AddBoth(graphIdxRow, graphIdxCol)
	}
	// Get the result of the bimax Function
	return BiMax(G, U, V)
}

// func BiMaxBinaryVerticesC(uu, vv []int) *BiMaxResult {
// }

func BiMaxBinaryVertices(uu, vv []int) *BiMaxResult {
	if len(uu) != len(vv) {
		panic(fmt.Sprintf("len(uu): %d len(vv): %d must be equal", len(uu), len(vv)))
	}
	U, V := NewSetFromSlice(uu), NewSetFromSlice(vv)
	G := graph.New(U.Card() + V.Card())
	for i := 0; i < len(uu); i++ {
		G.AddBoth(uu[i], vv[i]+U.Card())
	}
	return BiMax(G, U, V)
}

// BiMaxResult represents the result returned from 'BiMax' as a set of boths
// rows and columns or as the set of two vertecies in a the maximal biclique of
// graph G.
type BiMaxResult struct {
	Rows, Cols *SetOp
}

// ToC converts the BiMaxResult to allocted data in C so for ease of use in
// returning results functions 'BiMaxBinaryMatrixC' and 'BiMaxBinaryVerticesC', that
// deal with returning to the C api.
func (r *BiMaxResult) ToC() (C.size_t, *C.longlong, C.size_t, *C.longlong) {
	// getSetValues64 returns the values of a set as []int64.
	getSetValues64 := func(set *SetOp) []int64 {
		result := make([]int64, 0, set.Card())
		set.Each(func(v int) (done bool) {
			result = append(result, int64(v))
			return
		})
		return result
	}
	// Get the values of row and column sets in the result as []int64.
	rows := getSetValues64(r.Rows)
	cols := getSetValues64(r.Cols)
	// toArrC convert slice of int64to a pointer to allocated C.longlong memory
	// and the size of length of the allocated memory measured in C.longlong.
	toArrC := func(x []int64) (C.size_t, *C.longlong) {
		bb := *(*[]byte)(unsafe.Pointer(&x))
		bbH := (*reflect.SliceHeader)(unsafe.Pointer(&bb))
		bbH.Len = len(x) * int(unsafe.Sizeof(x[0]))
		return C.size_t(len(x)), (*C.longlong)(C.CBytes(bb))
	}
	// Get the values of rows and cols as C memory pointers to C.longlong data and
	// lengths of said pointers to data
	lenRowsC, dataRowsC := toArrC(rows)
	lenColsC, dataColsC := toArrC(cols)

	return lenRowsC, dataRowsC, lenColsC, dataColsC
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
		return G.Degree(v1) <= G.Degree(v2)
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
