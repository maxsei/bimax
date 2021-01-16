package main

import (
	"fmt"
	"log"

	"github.com/yourbasic/graph"
	"gorgonia.org/tensor"
)

const offsetBin int = 2

type Mat [][]uint8

func (m Mat) String() string {
	const matChar rune = 0x000023a1
	result := fmt.Sprintf("Matrix (%d, %d)\n", len(m), len(m[0]))
	for i, row := range m {
		// offset is 0 at start 1 during iteration, and 2 at the last iteration
		offset := rune(1 + (i-(i+1))/(i+1) + (i / (len(m) - 1)))
		rowRunes := []rune(fmt.Sprint(row))
		rowRunes[0] = matChar + offset
		rowRunes[len(rowRunes)-1] = matChar + offset + 3
		result += string(rowRunes) + "\n"
	}
	return result
}

func main() {
	// n, m := 10, 5
	// data := []uint8{
	// 	1, 0, 1, 0, 0,
	// 	0, 1, 0, 0, 1,
	// 	1, 1, 1, 0, 0,
	// 	0, 0, 0, 1, 0,
	// 	1, 1, 1, 1, 0,
	// 	0, 0, 0, 1, 1,
	// 	1, 1, 0, 0, 0,
	// 	0, 0, 1, 0, 1,
	// 	0, 0, 1, 0, 0,
	// 	1, 0, 0, 1, 0,
	// }
	n, m := 6, 8
	data := []uint8{
		1, 1, 1, 0, 1, 0, 1, 0,
		0, 0, 1, 1, 1, 0, 0, 1,
		0, 0, 0, 1, 1, 0, 0, 1,
		0, 0, 0, 1, 0, 1, 0, 1,
		0, 0, 1, 1, 0, 0, 0, 0,
		0, 1, 0, 1, 0, 1, 0, 1,
	}

	// n, m := 4, 6
	// data := []uint8{
	// 	0, 0, 1, 0, 0, 0,
	// 	1, 1, 1, 0, 0, 0,
	// 	0, 1, 0, 1, 1, 1,
	// 	1, 1, 0, 0, 0, 0,
	// }
	// n, m := 4, 4
	// data := []uint8{
	// 	0, 0, 1, 0,
	// 	1, 1, 1, 0,
	// 	0, 1, 0, 1,
	// 	1, 1, 0, 0,
	// }

	t := make(Mat, len(data)/m)
	for i := range t {
		t[i] = data[i*m : (i+1)*m]
	}
	fmt.Printf("t = %+v\n", t)

	// Create Range Table
	rt := NewRangeTable()
	// appendRangeLen(rt, "binary", 2)
	appendRangeLen(rt, "rows", n)
	appendRangeLen(rt, "columns", m)
	// appendRangeLen(rt, "indicies", t.Len())
	fmt.Printf("rt = %+v\n", rt)
	fmt.Println()

	// TODO: maybe find where 1's are is faster to do first and then using a
	// masked iterator <14-01-21, Max Schulte> //

	g := graph.New(n + m)
	var row, col int
	// Fill in Graph
	for i := 0; i < len(data); i++ {
		if data[i] == uint8(0) {
			continue
		}
		row = i / m
		col = i % m
		g.AddBoth(row, n+col)
	}

	// // Visit intersection nodes and show which edges they point to
	// for _, v := range []int{4} {
	// 	g.Visit(v, func(w int, c int64) (skip bool) {
	// 		// v1 := fmt.Sprintf("%d", w)
	// 		// v2 := fmt.Sprintf("%d", v)
	// 		// fmt.Printf("%d-> %d\n", v, w)
	// 		return false
	// 	})
	// }

	BiMax(g)

	// Save graph as graph and bigraph
	if err := SaveGraphiz("graph2.dot", g, rt, false); err != nil {
		log.Fatal(err)
	}
	if err := SaveGraphiz("bigraph2.dot", g, rt, true); err != nil {
		log.Fatal(err)
	}
}

func Conquer(g *graph.Mutable) ([]int, []int) {
	var conquer func([]int, []int, *graph.Mutable) ([]int, []int)
	depth := 0

	conquer = func(rows, cols []int, g *graph.Mutable) ([]int, []int) {
		depth++
		// Take first row in rows and shrink rows by 1
		row := rows[0]
		// otherRows := rows[1:]

		// Visit all nodes in that row and see if they are 1's
		g.Visit(row, func(u int, _ int64) (skip bool) {
			found := 0
			// Find where row == 1
			if g.Edge(u, 1) {
				found = 1
			}
			fmt.Printf("g[%d] idx %d is %d\n", u, u-10, found)
			return false
		})

		return []int{}, []int{}
	}

	// TODO: put t.shape[0] first here <14-01-21, Max Schulte> //
	rows := tensor.Range(tensor.Int, offsetBin, 4).([]int)
	cols := tensor.Range(tensor.Int, rows[len(rows)-1]+1, rows[len(rows)-1]+4+1).([]int)

	return conquer(rows, cols, g)
}

func PrettyPrintMapSymbols(sl []int, rt *RangeTable) {
	mapping := make(map[string][][2]int)
	for _, v := range sl {
		name, found := rt.Search(v)
		nameIdx := -1
		if found {
			r, _ := rt.RangeByName(name)
			nameIdx = v - r.Start()
		}
		vals, ok := mapping[name]
		if !ok {
			vals = [][2]int{}
		}
		vals = append(vals, [2]int{v, nameIdx})
		mapping[name] = vals
	}
	for _, name := range rt.Names() {
		pairs := mapping[name]
		fmt.Printf("%s = \n", name)
		for _, pair := range pairs {
			fmt.Printf("%d-> %d\n", pair[0], pair[1])
		}
	}
}

func appendRangeLen(rt *RangeTable, name string, length int) {
	if rt.Len() == 0 {
		if err := rt.Insert(name, Range{0, length}); err != nil {
			panic(err)
		}
		return
	}
	r := rt.Ranges()[0]
	if err := rt.Insert(name, Range{r.Stop(), r.Stop() + length}); err != nil {
		panic(err)
	}
	return
}
