package Transition

import (
	"bytes"
	. "chukuparser/Algorithm/FeatureVector"
	"chukuparser/Util"
	"fmt"
	"strings"
	"text/tabwriter"
)

type FeaturesList struct {
	Features   []Feature
	Transition Transition
	Previous   *FeaturesList
}

func (l *FeaturesList) String() string {
	var (
		retval []string      = make([]string, 0, 100)
		cur    *FeaturesList = l
	)
	for cur != nil {
		retval = append(retval, fmt.Sprintf("%v", cur.Transition))
		for _, val := range cur.Features {
			retval = append(retval, fmt.Sprintf("\t%v", val))
		}
		cur = cur.Previous
	}
	return strings.Join(retval, "\n")
}

type Transition int

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

	TransitionTypes() []string

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
