package Morph

import (
	"chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
	// "log"
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

func (m *BasicMorphGraph) GetMorpheme(i int) *NLP.EMorpheme {
	return m.Nodes[i].(*NLP.EMorpheme)
}

func (m *BasicMorphGraph) Sentence() NLP.Sentence {
	return m.Lattice
}

func (m *BasicMorphGraph) TaggedSentence() NLP.TaggedSentence {
	sent := make([]NLP.TaggedToken, m.NumberOfNodes()-1)
	for _, node := range m.Nodes {
		taggedNode := node.(*NLP.Morpheme)
		if taggedNode.Form == NLP.ROOT_TOKEN {
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

func CombineToGoldMorph(graph NLP.LabeledDependencyGraph, goldLat, ambLat NLP.LatticeSentence) (*BasicMorphGraph, bool) {
	var addedMissingSpellout bool
	// generate graph
	mGraph := new(Transition.BasicDepGraph)

	mGraph.Nodes = make([]NLP.DepNode, 0, graph.NumberOfNodes())

	// generate morph. disambiguation (= mapping) and nodes
	mappings := make([]*NLP.Mapping, len(goldLat))
	for i, lat := range goldLat {
		lat.GenSpellouts()
		lat.GenToken()
		mapping := &NLP.Mapping{
			lat.Token,
			lat.Spellouts[0],
		}
		// if the gold spellout doesn't exist in the lattice, add it
		_, exists := ambLat[i].Spellouts.Find(mapping.Spellout)
		if !exists {
			ambLat[i].Spellouts = append(ambLat[i].Spellouts, mapping.Spellout)
			addedMissingSpellout = true
		}

		mappings[i] = mapping

		// add the morpheme as a node
		for _, morpheme := range mapping.Spellout {
			mGraph.Nodes = append(mGraph.Nodes, morpheme)
		}
	}

	// copy arcs
	mGraph.Arcs = make([]*Transition.BasicDepArc, graph.NumberOfArcs())
	for i, arcId := range graph.GetEdges() {
		arc := graph.GetLabeledArc(arcId)
		// TODO: fix this ugly casting
		mGraph.Arcs[i] = arc.(*Transition.BasicDepArc)
	}

	m := &BasicMorphGraph{
		*mGraph,
		mappings,
		ambLat,
	}
	return m, addedMissingSpellout
}
