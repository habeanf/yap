package Morph

import (
	"chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
)

type BasicMorphGraph struct {
	Transition.BasicDepGraph
	Mappings []*NLP.Mapping
	Lattice  NLP.LatticeSentence
}

var _ NLP.MorphDependencyGraph = &BasicMorphGraph{}

func (m *BasicMorphGraph) GetMappings() []*NLP.Mapping {
	return m.Mappings
}

func (m *BasicMorphGraph) GetMorpheme(i int) *NLP.Morpheme {
	return m.Nodes[i].(*NLP.Morpheme)
}

func (m *BasicMorphGraph) Sentence() NLP.Sentence {
	return m.Lattice
}

func (m *BasicMorphGraph) TaggedSentence() NLP.TaggedSentence {
	sent := make([]NLP.TaggedToken, m.NumberOfNodes()-1)
	for _, node := range m.Nodes {
		taggedNode := node.(*NLP.Morpheme)
		if taggedNode.Form == Transition.ROOT_TOKEN {
			continue
		}
		target := taggedNode.ID() - 1
		if target < 0 {
			panic("Too small")
		}
		if target >= len(sent) {
			panic("Too large")
		}
		sent[target] = NLP.TaggedToken{taggedNode.Form, taggedNode.POS}
	}
	return NLP.TaggedSentence(NLP.BasicTaggedSentence(sent))
}
