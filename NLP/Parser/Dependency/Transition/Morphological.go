package Transition

type Morpheme string

type Spellout []Morpheme

type MorphologyAnalysis func(t *string) []Spellout

type Morphological interface {
	SetMorph(m *MorphologyAnalysis)
	GetMorph(m *MorphologyAnalysis)
}

type MRLParser struct {
	MA MorphologyAnalysis
}
