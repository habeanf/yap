package disambig

import (
	. "chukuparser/nlp/types"
)

type MorphologicalDisambiguator interface {
	Parse(LatticeSentence) (Mappings, interface{})
}
