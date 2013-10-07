package transition

import (
	"chukuparser/algorithm/perceptron"
	"chukuparser/nlp/parser/dependency"

	BeamSearch "chukuparser/algorithm/search"
)

type VarBeam struct {
	Beam
}

var _ BeamSearch.Interface = &VarBeam{}
var _ perceptron.EarlyUpdateInstanceDecoder = &VarBeam{}
var _ dependency.DependencyParser = &VarBeam{}

type NoCandidate struct{}

var _ BeamSearch.Candidate = &NoCandidate{}

func (c *NoCandidate) Copy() BeamSearch.Candidate {
	return c
}

func (c *NoCandidate) Score() float64 {
	return 0
}

func (c *NoCandidate) Equal(other BeamSearch.Candidate) bool {
	_, ok := other.(*NoCandidate)
	return ok
}

func (v *VarBeam) Top(a BeamSearch.Agenda) BeamSearch.Candidate {
	agenda := a.(*Agenda)
	for _, conf := range agenda.Confs {
		if !conf.C.Conf().Terminal() {
			return &NoCandidate{}
		}
	}
	return v.Beam.Top(a)
}

func (v *VarBeam) GoalTest(p BeamSearch.Problem, c BeamSearch.Candidate) bool {
	_, isNoCandidate := c.(*NoCandidate)
	if isNoCandidate {
		return false
	} else {
		return v.Beam.GoalTest(p, c)
	}
}
