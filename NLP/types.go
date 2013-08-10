package NLP

import "chukuparser/Util"

type Token string

type TaggedToken struct {
	Token string
	POS   string
}

type Sentence interface {
	Util.Equaler
	Tokens() []string
}

type TaggedSentence interface {
	Sentence
	TaggedTokens() []TaggedToken
}
