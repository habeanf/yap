package Conll

import "io"

type Row struct {
	ID      int
	Form    string
	CPosTag string
	PosTag  string
	Feats   map[string]string
	Head    int
	DepRel  string
}

type FileRow struct {
	ID      int
	Form    string
	Lemma   string
	CPosTag string
	PosTag  string
	Feats   string
	Head    int
	DepRel  string
	PHead   int
	PDepRel string
}

type Sentence map[int]ConllRow

func Read(r *io.Reader) []ConllSentence {
	var (
		sentences   []ConllSentence
		currentSent []ConllRow
		currentRow  string
	)

}

func Write(sentences []ConllSentence, w *io.Writer) {
	var (
		sentences   []ConllSentence
		currentSent []ConllRow
		currentRow  string
	)

}
