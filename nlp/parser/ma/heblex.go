package ma

import (
	. "yap/nlp/types"
)

type BGULex struct {
	MaxPrefixLen int
	Prefixes     map[string]BasicMorphemes

	Lex map[string]BasicMorphemes

	Files []string
}

var _ MorphologicalAnalyzer = &BGULex{}

func (l *BGULex) LoadPrefixes(file string) {

}

func (l *BGULex) LoadLex(file string) {

}

func (l *BGULex) Analyze(input []string) (LatticeSentence, interface{}) {
	retval := make(LatticeSentence, len(input))
	return retval, nil
}
