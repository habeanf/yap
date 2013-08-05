package Transition

import (
	"chukuparser/Algorithm/Graph"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
	"chukuparser/Util"
)

type DependencyConfiguration interface {
	Transition.Configuration
	NLP.LabeledDependencyGraph
	Address(location []byte) (int, bool)
	Attribute(nodeID int, attribute []byte) (string, bool)
	Previous() *DependencyConfiguration
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

type BasicDepGraph struct {
	Nodes []*NLP.DepNode
	Arcs  []*BasicDepArc
}

// Verify BasicDepGraph is a labeled dep. graph
var _ NLP.DependencyGraph = BasicDepGraph{}
var _ NLP.Labeled = BasicDepGraph{}

func (g BasicDepGraph) GetVertices() []int {
	return Util.RangeInt(len(g.Nodes))
}

func (g BasicDepGraph) GetEdges() []int {
	return Util.RangeInt(len(g.Edges))
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
