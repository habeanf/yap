package NLP

import (
	. "chukuparser/Algorithm/Graph"
)

type DepNode interface {
	Vertex
	String() string
}

type DepArc interface {
	DirectedEdge
	GetModifier() int
	GetHead() int
}

type DepRel string

type LabeledDepArc interface {
	DepArc
	GetRelation() DepRel
}

type Labeled interface {
	GetLabeledArc(int) *LabeledDepArc
}

type DependencyGraph interface {
	DirectedGraph
	GetNode(int) *DepNode
	GetArc(int) *DepArc
	NumberOfNodes() int
	NumberOfArcs() int
}

type LabeledDependencyGraph interface {
	DependencyGraph
	Labeled
}
