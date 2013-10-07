package Types

import (
	"chukuparser/algorithm/graph"
	"chukuparser/util"
)

type DepNode interface {
	Graph.Vertex
	String() string
}

type DepArc interface {
	Graph.DirectedEdge
	GetModifier() int
	GetHead() int
	String() string
}

type DepRel string

func (d DepRel) String() string {
	return string(d)
}

type LabeledDepArc interface {
	DepArc
	GetRelation() DepRel
}

type Labeled interface {
	GetLabeledArc(int) LabeledDepArc
}

type DependencyGraph interface {
	Graph.DirectedGraph
	GetNode(int) DepNode
	GetArc(int) DepArc
	NumberOfNodes() int
	NumberOfArcs() int
	Equal(otherEq Util.Equaler) bool
	Sentence() Sentence
	TaggedSentence() TaggedSentence
}

type LabeledDependencyGraph interface {
	DependencyGraph
	Labeled
}
