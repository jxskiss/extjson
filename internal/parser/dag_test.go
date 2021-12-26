package parser

import (
	"log"
	"reflect"
	"sort"
	"testing"
)

func TestDAG(t *testing.T) {
	var dag dag
	edges := [][2]int{
		{1, 2},
		{1, 3},
		{1, 4},
		{1, 5},
		{2, 6},
		{3, 5},
		{4, 7},
		{5, 6},
		{5, 7},
		{6, 8},
		{7, 8},
		{10, 11},
		{12, 11},
	}
	for _, e := range edges {
		dag.addEdge(e[0], e[1])
	}
	dag.addVertex(15)
	order := dag.sort()
	log.Println(order)
	assertTrue(t, isEqualOrder(order[:4], []int{1, 10, 12, 15}))
	assertTrue(t, isEqualOrder(order[4:8], []int{2, 3, 4, 11}))
	assertTrue(t, isEqualOrder(order[8:9], []int{5}))
	assertTrue(t, isEqualOrder(order[9:11], []int{6, 7}))
	assertTrue(t, order[len(order)-1] == 8)

	assertTrue(t, dag.isCyclic(5, 1))
	assertTrue(t, dag.isCyclic(7, 1))
	assertTrue(t, dag.isCyclic(8, 4))
}

func isEqualOrder(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Ints(a)
	sort.Ints(b)
	return reflect.DeepEqual(a, b)
}

func assertTrue(t *testing.T, value bool) {
	if !value {
		t.Helper()
		t.Fatal("Should be true")
	}
}
