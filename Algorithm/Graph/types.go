package Graph

type Vertex interface {
	ID() int
}

type Edge interface {
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
