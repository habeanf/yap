package Types

import (
	"chukuparser/Util"
	"reflect"
)

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

type BasicTaggedSentence []TaggedToken

func (b BasicTaggedSentence) Tokens() []string {
	tokens := make([]string, len(b))
	for i, token := range b {
		tokens[i] = token.Token
	}
	return tokens
}

func (b BasicTaggedSentence) TaggedTokens() []TaggedToken {
	return []TaggedToken(b)
}

func (b BasicTaggedSentence) Equal(otherEq Util.Equaler) bool {
	asTagged := otherEq.(BasicTaggedSentence)
	return reflect.DeepEqual(b, asTagged)
}
