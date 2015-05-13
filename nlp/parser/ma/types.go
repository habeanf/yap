package ma

import . "yap/nlp/types"

type MorphologicalAnalyzer interface {
	Analyze(input []string) (LatticeSentence, interface{})
}
