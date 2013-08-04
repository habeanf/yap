package Transition

import (
	"chukuparser/Algorithm/Graph"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
)

type DependencyConfiguration interface {
	Transition.Configuration
	NLP.DependencyGraph
}

type TaggedDepNode struct {
	id    int
	Token string
	POS   string
}

var _ NLP.DepNode = TaggedDepNode{}

func (t TaggedDepNode) ID() int {
	return t.id
}

func (t TaggedDepNode) String() string {
	return t.Token
}

func (t TaggedDepNode) GetProperty(prop string) (string, bool) {
	switch prop {
	case "w":
		return t.Token(), true
	case "p":
		return t.POS, true
	default:
		return "", false
	}
}

type BasicDepArc struct {
	Modifier int
	Relation NLP.DepRel
	Head     int
}

var _ NLP.LabeledDepArc = BasicDepArc{}

func (arc BasicDepArc) Vertices() []int {
	return []int{arc.Head, arc.Modifier}
}

func (arc BasicDepArc) From() int {
	return arc.Modifier
}

func (arc BasicDepArc) To() int {
	return arc.Head
}

func (arc BasicDepArc) GetHead() int {
	return arc.Head
}

func (arc BasicDepArc) GetModifier() int {
	return arc.Modifier
}

func (arc BasicDepArc) GetRelation() NLP.DepRel {
	return arc.Relation
}

func (arc BasicDepArc) GetProperty(property string) (string, bool) {
	if property == "l" {
		return arc.Relation, true
	} else {
		return "", false
	}
}

type BasicDepGraph struct {
	Nodes []*NLP.DepNode
	Arcs  []*BasicDepArc
}

// Verify BasicDepGraph is a labeled dep. graph
var _ NLP.DependencyGraph = BasicDepGraph{}
var _ NLP.Labeled = BasicDepGraph{}

func (g BasicDepGraph) GetVertices() []int {
	retval := make([]int, len(g.Nodes))
	for i := 0; i < len(g.Nodes); i++ {
		retval[i] = i
	}
	return retval
}

func (g BasicDepGraph) GetEdges() []int {
	retval := make([]int, len(g.Edges))
	for i := 0; i < len(g.Edges); i++ {
		retval[i] = i
	}
	return retval
}

func (g BasicDepGraph) GetVertex(n int) *Graph.Vertex {
	return g.Nodes[n]
}

func (g BasicDepGraph) GetEdge(n int) *Graph.Edge {
	return g.Arcs[n]
}

func (g BasicDepGraph) GetDirectedEdge(n int) *Graph.DirectedEdge {
	return g.Arcs[n]
}

func (g BasicDepGraph) NumberOfVertices() int {
	return len(g.Nodes)
}

func (g BasicDepGraph) NumberOfEdges() int {
	return len(g.Arcs)
}

func (g BasicDepGraph) GetNode(n int) *NLP.DepNode {
	return g.Nodes[n]
}

func (g BasicDepGraph) GetArc(n int) *NLP.DepArc {
	return g.Arcs[n]
}

func (g BasicDepGraph) NumberOfNodes() int {
	return g.NumberOfVertices()
}

func (g BasicDepGraph) NumberOfArcs() int {
	return g.NumberOfEdges()
}

func (g BasicDepGraph) GetLabeledArc(n int) *NLP.LabeledDepArc {
	return g.Arcs[n]
}
