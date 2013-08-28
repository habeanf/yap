package Transition

import (
	"bytes"
	"chukuparser/Util"
	"text/tabwriter"
)

type Transition string

type Configuration interface {
	Init(interface{})
	Terminal() bool

	Copy() Configuration
	GetSequence() ConfigurationSequence
	SetLastTransition(Transition)
	GetLastTransition() Transition
	String() string
	Equal(otherEq Util.Equaler) bool
}

type ConfigurationSequence []Configuration

type TransitionSystem interface {
	Transition(from Configuration, transition Transition) Configuration

	TransitionTypes() []Transition

	YieldTransitions(conf Configuration) chan Transition

	Oracle() Oracle
	AddDefaultOracle()
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
		w.Write(asBytes)
		if i < seqLength-1 {
			w.Write([]byte{'\n'})
		}
	}
	w.Flush()
	return buf.String()
}

func (seq ConfigurationSequence) SharedTransitions(other ConfigurationSequence) int {
	lenOther := len(other)
	lenSeq := len(seq)
	sharedSeq := 0
	for i, _ := range seq {
		if len(other) <= i {
			break
		}
		if other[lenOther-i-1].GetLastTransition() != seq[lenSeq-i-1].GetLastTransition() {
			break
		}
		sharedSeq++
	}
	return sharedSeq
}

func (seq ConfigurationSequence) Equal(otherEq Util.Equaler) bool {
	other := otherEq.(ConfigurationSequence)
	for i, val := range seq {
		if !other[i].Equal(val) {
			return false
		}
	}
	return true
}
