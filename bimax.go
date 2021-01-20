package main

import (
	"fmt"

	"C"

	"github.com/yourbasic/graph"
)
import "unsafe"

// func main() {}

// func BiMaxBinaryMatrixC(n, m int, data []uint8) **C.GoInt {

//export BiMaxBinaryMatrixC
// func BiMaxBinaryMatrixC(n, m int, data []uint8) (C.size_t, C.size_t, **C.longlong) {
// func BiMaxBinaryMatrixC(n, m int, data []uint8) (C.size_t, **C.longlong) {
func BiMaxBinaryMatrixC(n, m int, data []uint8) (C.size_t, *C.longlong) {
	rows, cols := BiMaxBinaryMatrix(n, m, data)
	fmt.Printf("rows = %+v\n", rows)
	fmt.Printf("cols = %+v\n", cols)
	// p = C.malloc(C.size_t(len(result)) * C.size_t(unsafe.Sizeof(uintptr(0))))

	size := len(rows) + len(cols)
	// allocate the *C.double array
	dataC := C.malloc(C.size_t(size) * C.size_t(unsafe.Sizeof(int64(0))))

	// convert the pointer to a go slice so we can index it
	doubles := (*[1<<30 - 1]C.longlong)(dataC)[:size:size]
	offset := 0
	for i, vv := range [][]int{rows, cols} {
		for j, v := range vv {
			v64 := int64(v)
			doubles[i*offset+j] = *(*C.longlong)(unsafe.Pointer(&v64))
		}
		offset += len(vv)
	}

	// return C.size_t(len(result[0])), C.size_t(len(result[1])), (*C.double)(p)
	return C.size_t(size), (*C.longlong)(dataC)
}

// BiMaxBinaryMatrix takes in an n by m binary matrix where n is the number of
// rows and m the number of columns.  The data for the binary matrix is
// specified as a slice of uint8 containing only 1's and 0's.
//export BiMaxBinaryMatrix
func BiMaxBinaryMatrix(n, m int, data []uint8) ([]int, []int) {
	fmt.Printf("data = %+v\n", data)
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
	rows := rowset.Values()
	cols := colset.Values()
	// Subtract the number of rows from columns to go from graph column index to
	// column index
	for i := 0; i < len(cols); i++ {
		cols[i] -= n
	}
	return rows, cols
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
	P := PU.Order(func(v1, v2 int) bool {
		return G.Degree(v1) <= G.Degree(v2)
	})

	// Q: is a set of verticies used to determine maximality, initially empty
	Q := NewSet()

	// Resulting sets
	A, B := NewSet(), NewSet()

	var bicliqueFind func(P *OrderedSet, L, R, Q *UnorderedSet)
	bicliqueFind = func(P *OrderedSet, L, R, Q *UnorderedSet) {
		for i := 0; 0 < P.Card(); i++ {
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
			CValues := c.Values()
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
