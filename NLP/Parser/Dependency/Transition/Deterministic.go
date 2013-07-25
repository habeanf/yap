package Transition

import (
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP"
)

type Deterministic struct {
	transFunc *TransitionSystem
}

func (d *Deterministic) ParseOracle(sent Sentence, gold *DependencyGraph) (*DependencyGraph, []Configuration) {
	c := new(Configuration)
	c.Init(sent)
	d.transFunc.SetGold(gold)
	for !c.Terminal() {
		transition := d.transFunc.Oracle().GetTransition(c)
		c = d.transFunc.Transition(c, transition)
	}
	return c.Graph(), c.GetSequence()
}

func (d *Deterministic) Parse(sent Sentence, constraints *interface{}, model *interface{}) (*DependencyGraph, []Configuration) {
	if constraints != nil {
		warn("Got non-nil constraints; Note that deterministic dependency parsing does not consider constraints")
	}
	classifier, ok := model.(*Decision)
	if !ok {
		panic("Parameter model is not a Transition.Decision, cannot use as a classifier")
	}
	c := new(Configuration)
	c.Init(sent)
	for !c.Terminal() {
		transition := classifier.GetTransition(c)
		c = d.transFunc.Transition(c, transition)
	}
	return c.Graph(), c.GetSequence()
}
