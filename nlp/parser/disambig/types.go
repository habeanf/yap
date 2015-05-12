package disambig

import (
	. "yap/nlp/types"
)

type MorphologicalDisambiguator interface {
	Parse(LatticeSentence) (Mappings, interface{})
}
