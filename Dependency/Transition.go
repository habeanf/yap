package DepParser

type Parser interface {
	Init()
}

type Morpheme string

type Spellout []Morpheme

type MorphologyAnalysis func(t *Token) []Spellout

type Morphological interface {
	SetMorph(m *MorphologyAnalysis)
	GetMorph(m *MorphologyAnalysis)
}

type MRLParser struct {
	MA MorphologyAnalysis
}

type Configuration struct {
	Stack  interface{}
	Queue  interface{}
	Arcs   interface{}
	Morphs interface{}
}
