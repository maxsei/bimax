package main

import (
	"fmt"
	"os"

	"github.com/yourbasic/graph"

	"github.com/awalterschulze/gographviz"
)

func SaveGraphiz(filename string, g graph.Iterator, rt *RangeTable, bipart bool) error {
	// Open file with filename
	w, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer w.Close()

	// Create graphiz graph new graph
	graphAst, _ := gographviz.ParseString(`digraph G {}`)
	gg := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, gg); err != nil {
		return err
	}
	// Create sub graph and means for adding new subgrpahs
	subgraphs := gographviz.NewSubGraphs()
	numSubgraphs := 0
	getSubgraphName := func(x int) string { return fmt.Sprintf("%d", x) }
	appendSubGraph := func() func() { return func() { subgraphs.Add(getSubgraphName(numSubgraphs)); numSubgraphs++ } }()
	// Append single subgraph by default
	appendSubGraph()

	// Split graphs into subgraphs if needed
	var part []int
	if bipart {
		var ok bool
		part, ok = graph.Bipartition(g)
		if !ok {
			return fmt.Errorf("could not bipartition graph %v", g)
		}
		appendSubGraph()
	}

	// Set verticies
	var verticies [2][]int
	verticies[1] = part
	for v1 := 0; v1 < g.Order(); v1++ {
		isPart := false
		for _, v2 := range part {
			if v1 == v2 {
				isPart = true
				break
			}
		}
		if !isPart {
			verticies[0] = append(verticies[0], v1)
		}
	}

	// Visit all node edges
	for subgraph := 0; subgraph < numSubgraphs; subgraph++ {
		// Set subraph name
		subgraphName := getSubgraphName(subgraph)
		if !bipart {
			subgraphName = gg.Name
		}
		for _, v := range verticies[subgraph] {
			var err error
			// If bipart traverse the ith part instead
			g.Visit(v, func(u int, _ int64) (skip bool) {
				// edge strings and ints
				var edgeStrs [2]string
				edgeInts := [2]int{u, v}
				// Set edge strs
				for i, x := range edgeInts {
					edgeStrs[i] = fmt.Sprintf("%d", x)
				}
				// Set edge strs to what the edge int is in the RangeTable
				if rt != nil {
					for i, x := range edgeInts {
						name, ok := rt.Search(x)
						// If not found then return error and skip other visits
						if !ok {
							err = fmt.Errorf("name for vertex %d doesn't exists in %v", x, rt)
							return true
						}
						r, _ := rt.RangeByName(name)
						nameIdx := x - r.Start()
						edgeStrs[i] = fmt.Sprintf("_%d_%s_%d", x, name, nameIdx)
					}
				}

				// Add new edge to graphiz graph and skip other visits if error
				err = AddBoth(gg, subgraphName, edgeStrs[0], edgeStrs[1])
				if err != nil {
					return true
				}
				return false
			})
			if err != nil {
				return err
			}
		}
		if bipart {
			// Set subgraph
			if err := gg.AddSubGraph(gg.Name, subgraphName, map[string]string{"rank": "same"}); err != nil {
				return err
			}
		}
	}
	_, err = w.Write([]byte(gg.String()))
	return err
}

func AddBoth(gg *gographviz.Graph, parent, v1, v2 string) error {
	if err := gg.AddNode(parent, v2, nil); err != nil {
		return err
	}
	if err := gg.AddNode(parent, v2, nil); err != nil {
		return err
	}
	if err := gg.AddEdge(v1, v2, true, nil); err != nil {
		return err
	}
	return nil
}
