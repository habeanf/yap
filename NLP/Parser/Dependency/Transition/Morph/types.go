package Morph

import (
	"chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
)

type Mapping struct {
	Token    NLP.Token
	Spellout Spellout
}

type MorphConfiguration struct {
	Transition.SimpleConfiguration
	LatticeQueue Stack
	Lattices     []*NLP.Lattice
	Mappings     []*Mapping
	MorphNodes   []*Morpheme
}

// Verify that MorphConfiguration is a Configuration
var _ DependencyConfiguration = &MorphConfiguration{}
var _ NLP.DependencyGraph = &MorphConfiguration{}
