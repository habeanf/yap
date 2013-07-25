package Transition

import (
	"bytes"
	"text/tabwriter"
)

type Transition string

type Configuration interface {
	Init(interface{})
	Terminal() bool

	Copy() interface{}
	GetSequence() ConfigurationSequence
	SetLastTransition(Transition)
	String() string
}

type ConfigurationSequence []*Configuration

type TransitionSystem interface {
	Transition(from interface{}, transition Transition) *Configuration

	TransitionTypes() []Transition

	Oracle() *Decision
}

type Decision interface {
	GetTransition(*Configuration) Transition
}

func (seq ConfigurationSequence) String() string {
	var buf bytes.Buffer
	w := new(tabwriter.Writer)
	w.Init(&buf, 0, 8, 0, '\t', 0)
	seqLength := len(seq)
	for i, _ := range seq {
		conf := seq[seqLength-i-1]
		asString := (*conf).String()
		asBytes := []byte(asString)
		w.Write(asBytes)
	}
	w.Flush()
	return buf.String()
}
