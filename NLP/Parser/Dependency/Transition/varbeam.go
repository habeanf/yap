package Transition

import (
	"chukuparser/Algorithm/Perceptron"
	"chukuparser/NLP/Parser/Dependency"

	BeamSearch "chukuparser/Algorithm/Search"
)

type VarBeam struct {
	Beam
}

var _ BeamSearch.Interface = &VarBeam{}
var _ Perceptron.EarlyUpdateInstanceDecoder = &VarBeam{}
var _ Dependency.DependencyParser = &VarBeam{}

type NoCandidate struct{}

var _ BeamSearch.Candidate = &NoCandidate{}

func (c *NoCandidate) Copy() BeamSearch.Candidate {
	return c
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
