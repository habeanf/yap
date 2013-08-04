package Transition

import (
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP"
)

type Deterministic struct {
	transFunc *TransitionSystem
}

func (d *Deterministic) ParseOracle(sent Sentence, gold *DependencyGraph) (*DependencyGraph, ConfigurationSequence) {
	newConf := new(SimpleConfiguration)
	var c Configuration = *newConf
	c.Init(sent)
	transitionSystem := *(d.transFunc)
	oracle := *(transitionSystem.Oracle())
	goldGeneric := (interface{})(*gold)
	oracle.SetGold(&goldGeneric)
	for !c.Terminal() {
		transition := oracle.GetTransition(&c)
		c = *transitionSystem.Transition(&c, transition)
	}
	configurationAsGraph := c.(DependencyGraph)
	return &configurationAsGraph, c.GetSequence()
}

func (d *Deterministic) Parse(sent Sentence, constraints *interface{}, model *interface{}) (*DependencyGraph, ConfigurationSequence) {
	// if constraints != nil {
	// 	warn("Got non-nil constraints; Note that deterministic dependency parsing does not consider constraints")
	// }
	classifier, ok := (*model).(Decision)
	if !ok {
		panic("Parameter model is not a Transition.Decision, cannot use as a classifier")
	}
	c := (*(new(Configuration)))
	c.Init(sent)
	transitionSystem := *(d.transFunc)
	for !c.Terminal() {
		transition := classifier.GetTransition(&c)
		c = *(transitionSystem.Transition(&c, transition))
	}
	configurationAsGraph := c.(DependencyGraph)
	return &configurationAsGraph, c.GetSequence()
}
