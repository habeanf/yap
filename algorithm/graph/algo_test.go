package graph

import (
	"log"
	"reflect"
	"testing"
)

func TestYieldAllPaths(t *testing.T) {
	vertices := []BasicVertex{1, 2, 3, 4, 5, 6}
	edges := []BasicDirectedEdge{
		{1, 2},
		{1, 3},
		{2, 4},
		{4, 5},
		{4, 6},
		{5, 6},
		{3, 6},
		{1, 6},
	}
	graph := &BasicGraph{vertices, edges}
	paths := make([][]int, 0, 1)
	for path := range YieldAllPaths(graph, 1, 6) {
		// log.Println("Got path:", path)
		paths = append(paths, path)
	}
	shouldEq := [][]int{
		{1, 6}, {1, 3, 6}, {1, 2, 4, 6}, {1, 2, 4, 5, 6},
	}
	log.Println("Should be\n", shouldEq)
	log.Println("Got\n", paths)
	if !reflect.DeepEqual(paths, shouldEq) {
		t.Error("All paths not found")
	}
}
