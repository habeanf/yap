package Graph

import "chukuparser/Util"

type Vertex interface {
	Util.Equaler
	ID() int
}

type Edge interface {
	Util.Equaler
	Vertices() []int
}

type DirectedEdge interface {
	Edge
	From() int
	To() int
}

type Graph interface {
	GetVertices() []int
	GetEdges() []int
	GetVertex(int) Vertex
	GetEdge(int) Edge
	NumberOfVertices() int
	NumberOfEdges() int
}

type DirectedGraph interface {
	Graph
	GetDirectedEdge(int) DirectedEdge
}

type Lattice interface {
	Inf(int, int) int
	Sup(int, int) int
}

type BoundedLattice interface {
	Lattice
	Top() int
	Bottom() int
}

type BasicDirectedEdge [2]int

var _ DirectedEdge = BasicDirectedEdge{}

func (e BasicDirectedEdge) From() int {
	return e[0]
}

func (e BasicDirectedEdge) To() int {
	return e[1]
}

func (e BasicDirectedEdge) Vertices() []int {
	return []int{e[0], e[1]}
}

func (e BasicDirectedEdge) Equal(otherEq Util.Equaler) bool {
	other := otherEq.(BasicDirectedEdge)
	return e[0] == other[0] && e[1] == other[1]
}
