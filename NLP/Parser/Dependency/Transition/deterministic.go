package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
	"chukuparser/NLP/Parser/Dependency"
)

type Deterministic struct {
	TransFunc        Transition.TransitionSystem
	ReturnModelValue bool
	ReturnSequence   bool
}

var _ Dependency.DependencyParser = &Deterministic{}

type ParseResultParameters struct {
	modelValue interface{}
	sequence   Transition.ConfigurationSequence
}

var _ Perceptron.InstanceDecoder = &Deterministic{}

// Parser functions
func (d *Deterministic) Parse(sent NLP.Sentence, constraints Dependency.ConstraintModel, model Dependency.ParameterModel) (NLP.DependencyGraph, interface{}) {
	if constraints != nil {
		panic("Got non-nil constraints; deterministic dependency parsing does not consider constraints")
	}
	model, ok := model.(Perceptron.Model)
	if !ok {
		panic("Parameter model is not a Transition.Decision, cannot use as a classifier")
	}

	classifier := Transition.Decision(TransitionClassifier{model})
	c := Transition.Configuration(new(SimpleConfiguration))

	// deterministic parsing algorithm
	c.Init(sent)
	for !c.Terminal() {
		c, _ := classifier.Transition(c)
	}

	// build result parameters
	var resultParams *ParseResultParameters
	if d.ReturnModelValue || d.ReturnSequence {
		resultParams = new(ParseResultParameters)
		if d.ReturnModelValue {
			resultParams.modelValue = classifier.Weights
		}
		if d.ReturnSequence {
			resultParams.sequence = c.GetSequence()
		}
	}
	configurationAsGraph := c.(NLP.DependencyGraph)
	return configurationAsGraph, resultParams
}

// Perceptron functions
func (d *Deterministic) Decode(instance Perceptron.Instance, m Perceptron.Model) (Perceptron.DecodedInstance, *Perceptron.SparseWeightVector) {
	sent := instance.(NLP.Sentence)
	model := m.(Dependency.ParameterModel)
	d.ReturnModelValue = true
	graph, parseParamsInterface := d.Parse(sent, nil, model)
	parseParams := parseParamsInterface.(ParseResultParameters)
	weights := parseParams.modelValue.(*Perceptron.SparseWeightVector)
	return &Perceptron.Decoded{instance, graph}, weights
}

func (d *Deterministic) GoldDecode(goldInstance Perceptron.DecodedInstance, m Perceptron.Model) (Perceptron.DecodedInstance, *Perceptron.SparseWeightVector) {
	sent := goldInstance.Instance().(NLP.Sentence)
	model := m.(Dependency.ParameterModel)
	graph := goldInstance.Decoded().(NLP.DependencyGraph)
	d.ReturnModelValue = true
	parsedGraph, parseParamsInterface := d.ParseOracle(sent, graph, nil, model)
	if !graph.Equal(parsedGraph) {
		panic("Oracle result did not equal gold")
	}
	parseParams := parseParamsInterface.(ParseResultParameters)
	weights := parseParams.modelValue.(*Perceptron.SparseWeightVector)
	return &Perceptron.Decoded{goldInstance, graph}, weights
}

func (d *Deterministic) ParseOracle(sent NLP.Sentence, gold NLP.DependencyGraph, constraints interface{}, model interface{}) (NLP.DependencyGraph, interface{}) {
	if constraints != nil {
		panic("Got non-nil constraints; deterministic dependency parsing does not consider constraints")
	}
	c := Transition.Configuration(new(SimpleConfiguration))
	classifier := TransitionClassifier{model}
	c.Init(sent)
	oracle := d.TransFunc.Oracle()
	oracle.SetGold(gold)
	for !c.Terminal() {
		c, transition := oracle.Transition(c)
		classifier.Transition(transition)
	}

	// build result parameters
	var resultParams *ParseResultParameters
	if d.ReturnModelValue || d.ReturnSequence {
		resultParams = new(ParseResultParameters)
		if d.ReturnModelValue {
			resultParams.modelValue = classifier.Weights
		}
		if d.ReturnSequence {
			resultParams.sequence = c.GetSequence()
		}
	}
	configurationAsGraph := c.(NLP.DependencyGraph)
	return configurationAsGraph, resultParams
}

type TransitionClassifier struct {
	Model      Dependency.ParameterModel
	TransFunc  Transition.TransitionSystem
	ModelValue interface{}
}

func (tc *TransitionClassifier) Transition(c Transition.Configuration) (Transition.Configuration, Transition.Transition) {
	var (
		bestScore             float64
		bestConf, currentConf Transition.Configuration
		bestTransition        Transition.Transition
		tChan                 chan Transition.Transition = make(chan Transition.Transition)
	)
	go tc.TransFunc.PossibleTransitions(c, tChan)
	for transition := range tChan {
		currentConf = tc.TransFunc.Transition(c, transition)
		if currentScore, modelValue := tc.Model.Score(currentConf); currentScore > bestScore {
			bestScore, bestConf, bestTransition = currentScore, currentConf, transition
		}
	}
	return bestConf, bestTransition
}
