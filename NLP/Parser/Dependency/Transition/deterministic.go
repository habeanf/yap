package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP"
)

type Deterministic struct {
	transFunc TransitionSystem
}

var _ Perceptron.InstanceDecoder = &Deterministic{}

func (d *Deterministic) Decode(instance Perceptron.Instance, m Perceptron.Model) Perceptron.DecodedInstance {
	sent := instance.(Sentence)
	graph, _ := d.Parse(sent, nil, m)
	return Perceptron.Decoded{instance, graph}
}

func (d *Deterministic) Parse(sent Sentence, constraints interface{}, model interface{}) (DependencyGraph, ConfigurationSequence) {
	if constraints != nil {
		panic("Got non-nil constraints; deterministic dependency parsing does not consider constraints")
	}
	classifier, ok := model.(Decision)
	if !ok {
		panic("Parameter model is not a Transition.Decision, cannot use as a classifier")
	}
	c := Configuration(new(SimpleConfiguration))
	c.Init(sent)
	for !c.Terminal() {
		transition := classifier.GetTransition(c)
		c = d.transFunc.Transition(c, transition)
	}
	configurationAsGraph := c.(DependencyGraph)
	return configurationAsGraph, c.GetSequence()
}

func (d *Deterministic) Train() {

}

func (d *Deterministic) ParseOracle(sent Sentence, gold DependencyGraph) (DependencyGraph, ConfigurationSequence) {
	c := Configuration(new(SimpleConfiguration))
	c.Init(sent)
	oracle := d.transFunc.Oracle()
	oracle.SetGold(gold)
	for !c.Terminal() {
		transition := oracle.GetTransition(c)
		c = d.transFunc.Transition(c, transition)
	}
	configurationAsGraph := c.(DependencyGraph)
	return configurationAsGraph, c.GetSequence()
}
