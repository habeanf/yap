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
	Util.Equaler
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
