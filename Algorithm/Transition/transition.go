package Transition

import (
	"bytes"
	"text/tabwriter"
)

type Transition string

type Configuration interface {
	Init(interface{})
	Terminal() bool

	Copy() Configuration
	GetSequence() ConfigurationSequence
	SetLastTransition(Transition)
	String() string
}

type ConfigurationSequence []Configuration

type TransitionSystem interface {
	Transition(from Configuration, transition Transition) Configuration

	TransitionTypes() []Transition

	PossibleTransitions(conf Configuration, transitions chan Transition)

	Oracle() Oracle
}

type Decision interface {
	Transition(Configuration) Transition
}

type Oracle interface {
	Decision
	SetGold(interface{})
}

func (seq ConfigurationSequence) String() string {
	var buf bytes.Buffer
	w := new(tabwriter.Writer)
	w.Init(&buf, 0, 8, 0, '\t', 0)
	seqLength := len(seq)
	for i, _ := range seq {
		conf := seq[seqLength-i-1]
		asString := conf.String()
		asBytes := []byte(asString)
		w.Write(append(asBytes, '\n'))
	}
	w.Flush()
	return buf.String()
}
