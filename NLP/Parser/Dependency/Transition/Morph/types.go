package Morph

import (
	"chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
)

type BasicMorphGraph struct {
	Transition.BasicDepGraph
	Mappings []*NLP.Mapping
}

var _ NLP.MorphDependencyGraph = &BasicMorphGraph{}

func (m *BasicMorphGraph) GetMappings() []*NLP.Mapping {
	return m.Mappings
}

func (m *BasicMorphGraph) GetMorpheme(i int) *NLP.Morpheme {
	return m.Nodes[i].(*NLP.Morpheme)
}
