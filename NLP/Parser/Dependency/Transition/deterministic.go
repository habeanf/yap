package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
	"chukuparser/NLP/Parser/Dependency"
	"chukuparser/Util"
	"fmt"
	"sort"
)

type Deterministic struct {
	TransFunc          Transition.TransitionSystem
	FeatExtractor      Perceptron.FeatureExtractor
	ReturnModelValue   bool
	ReturnSequence     bool
	ShowConsiderations bool
}

var _ Dependency.DependencyParser = &Deterministic{}
var _ Perceptron.InstanceDecoder = &Deterministic{}

type ParseResultParameters struct {
	modelValue interface{}
	Sequence   Transition.ConfigurationSequence
}

// Parser functions
func (d *Deterministic) Parse(sent NLP.Sentence, constraints Dependency.ConstraintModel, model Dependency.ParameterModel) (NLP.DependencyGraph, interface{}) {
	if constraints != nil {
		panic("Got non-nil constraints; deterministic dependency parsing does not consider constraints")
	}
	if d.TransFunc == nil {
		panic("Can't parse without a transition system")
	}
	transitionClassifier := &TransitionClassifier{Model: model, TransFunc: d.TransFunc, FeatExtractor: d.FeatExtractor}
	transitionClassifier.Init()
	transitionClassifier.ShowConsiderations = d.ShowConsiderations
	c := Transition.Configuration(new(SimpleConfiguration))

	// deterministic parsing algorithm
	c.Init(sent)
	for !c.Terminal() {
		c, _ = transitionClassifier.TransitionWithConf(c)
		transitionClassifier.Increment(c)
		if c == nil {
			fmt.Println("Got nil configuration!")
		}
	}

	// build result parameters
	var resultParams *ParseResultParameters
	if d.ReturnModelValue || d.ReturnSequence {
		resultParams = new(ParseResultParameters)
		if d.ReturnModelValue {
			resultParams.modelValue = transitionClassifier.ModelValue
		}
		if d.ReturnSequence {
			resultParams.Sequence = c.GetSequence()
		}
	}
	configurationAsGraph := c.(NLP.DependencyGraph)
	return configurationAsGraph, resultParams
}

func (d *Deterministic) ParseOracle(gold NLP.DependencyGraph, constraints interface{}, model Dependency.ParameterModel) (NLP.DependencyGraph, interface{}) {
	if constraints != nil {
		panic("Got non-nil constraints; deterministic dependency parsing does not consider constraints")
	}
	if d.TransFunc == nil {
		panic("Can't parse without a transition system")
	}
	sent := gold.TaggedSentence()
	c := Transition.Configuration(new(SimpleConfiguration))
	c.Init(sent)
	classifier := TransitionClassifier{Model: model, FeatExtractor: d.FeatExtractor, TransFunc: d.TransFunc}

	classifier.Init()
	oracle := d.TransFunc.Oracle()
	oracle.SetGold(gold)
	for !c.Terminal() {
		transition := oracle.Transition(c)
		c = d.TransFunc.Transition(c, transition)
		classifier.Increment(c)
	}

	// build result parameters
	var resultParams *ParseResultParameters
	if d.ReturnModelValue || d.ReturnSequence {
		resultParams = new(ParseResultParameters)
		if d.ReturnModelValue {
			resultParams.modelValue = classifier.ModelValue
		}
		if d.ReturnSequence {
			resultParams.Sequence = c.GetSequence()
		}
	}
	configurationAsGraph := c.(*SimpleConfiguration).Graph()
	return configurationAsGraph, resultParams
}

func (d *Deterministic) ParseOracleEarlyUpdate(gold NLP.DependencyGraph, constraints interface{}, model Dependency.ParameterModel) (NLP.DependencyGraph, interface{}, interface{}) {
	if constraints != nil {
		panic("Got non-nil constraints; deterministic dependency parsing does not consider constraints")
	}
	if d.TransFunc == nil {
		panic("Can't parse without a transition system")
	}

	// Initializations
	sent := gold.TaggedSentence()

	c := Transition.Configuration(new(SimpleConfiguration))
	classifier := TransitionClassifier{Model: model, FeatExtractor: d.FeatExtractor, TransFunc: d.TransFunc}
	classifier.ShowConsiderations = d.ShowConsiderations

	oracle := d.TransFunc.Oracle()
	oracle.SetGold(gold)

	c.Init(sent)
	classifier.Init()

	var (
		goldWeights, predCurrentWeights interface{}
		predTrans                       Transition.Transition
		predFeatures                    []Perceptron.Feature
	)
	predWeights := classifier.Model.NewModelValue()
	for !c.Terminal() {
		goldTrans := oracle.Transition(c)
		goldConf := d.TransFunc.Transition(c, goldTrans)
		c, predTrans = classifier.TransitionWithConf(c)
		if c == nil {
			panic("Got nil configuration!")
		}

		predFeatures = d.FeatExtractor.Features(c)
		predCurrentWeights = classifier.Model.ModelValueOnes(predFeatures)

		// verify the right transition was chosen
		if predTrans != goldTrans {
			goldFeatures := d.FeatExtractor.Features(goldConf)
			goldCurrentWeights := classifier.Model.ModelValueOnes(goldFeatures)
			goldWeights = predWeights.ValueWith(goldCurrentWeights)
			predWeights = predWeights.ValueWith(predCurrentWeights)
			break
		}
		predWeights.Increment(predCurrentWeights)
	}

	// build result parameters
	var resultParams *ParseResultParameters
	if d.ReturnModelValue || d.ReturnSequence {
		resultParams = new(ParseResultParameters)
		if d.ReturnModelValue {
			resultParams.modelValue = predWeights
		}
		if d.ReturnSequence {
			resultParams.Sequence = c.GetSequence()
		}
	}
	configurationAsGraph := c.(*SimpleConfiguration).Graph()
	return configurationAsGraph, resultParams, goldWeights
}

// Perceptron functions
func (d *Deterministic) Decode(instance Perceptron.Instance, m Perceptron.Model) (Perceptron.DecodedInstance, *Perceptron.SparseWeightVector) {
	sent := instance.(NLP.Sentence)
	model := Dependency.ParameterModel(&PerceptronModel{m.(*Perceptron.LinearPerceptron)})
	d.ReturnModelValue = true
	graph, parseParamsInterface := d.Parse(sent, nil, model)
	parseParams := parseParamsInterface.(*ParseResultParameters)
	weights := parseParams.modelValue.(*PerceptronModelValue).vector
	return &Perceptron.Decoded{instance, graph}, weights
}

func (d *Deterministic) DecodeGold(goldInstance Perceptron.DecodedInstance, m Perceptron.Model) (Perceptron.DecodedInstance, *Perceptron.SparseWeightVector) {
	graph := goldInstance.Decoded().(NLP.DependencyGraph)
	model := Dependency.ParameterModel(&PerceptronModel{m.(*Perceptron.LinearPerceptron)})
	d.ReturnModelValue = true
	parsedGraph, parseParamsInterface := d.ParseOracle(graph, nil, model)
	if !graph.Equal(parsedGraph) {
		panic("Oracle parse result does not equal gold")
	}
	parseParams := parseParamsInterface.(*ParseResultParameters)
	weights := parseParams.modelValue.(*PerceptronModelValue).vector
	return &Perceptron.Decoded{goldInstance.Instance(), graph}, weights
}

func (d *Deterministic) DecodeEarlyUpdate(goldInstance Perceptron.DecodedInstance, m Perceptron.Model) (Perceptron.DecodedInstance, *Perceptron.SparseWeightVector, *Perceptron.SparseWeightVector) {
	graph := goldInstance.Decoded().(NLP.DependencyGraph)
	model := Dependency.ParameterModel(&PerceptronModel{m.(*Perceptron.LinearPerceptron)})
	d.ReturnModelValue = true
	parsedGraph, parseParamsInterface, goldWeights := d.ParseOracleEarlyUpdate(graph, nil, model)
	if parsedGraph.NumberOfEdges() == graph.NumberOfEdges() && !graph.Equal(parsedGraph) {
		panic("Oracle parse result does not equal gold")
	}
	parseParams := parseParamsInterface.(*ParseResultParameters)
	weights := parseParams.modelValue.(*PerceptronModelValue).vector
	return &Perceptron.Decoded{goldInstance.Instance(), parsedGraph}, weights, goldWeights.(*PerceptronModelValue).vector
}

type TransitionClassifier struct {
	Model              Dependency.ParameterModel
	TransFunc          Transition.TransitionSystem
	FeatExtractor      Perceptron.FeatureExtractor
	ModelValue         Dependency.ParameterModelValue
	ShowConsiderations bool
}

func (tc *TransitionClassifier) Init() {
	tc.ModelValue = tc.Model.NewModelValue()
}

func (tc *TransitionClassifier) Increment(c Transition.Configuration) *TransitionClassifier {
	features := tc.FeatExtractor.Features(Perceptron.Instance(c))
	modelValue := tc.Model.ModelValue(features)
	tc.ModelValue.Increment(modelValue)
	return tc
}

func (tc *TransitionClassifier) ScoreWithConf(c Transition.Configuration) float64 {
	features := tc.FeatExtractor.Features(Perceptron.Instance(c))
	weights := tc.Model.ModelValue(features)
	return tc.ModelValue.ValueWith(weights).Score()
}

func (tc *TransitionClassifier) Transition(c Transition.Configuration) Transition.Transition {
	_, transition := tc.TransitionWithConf(c)
	return transition
}

func (tc *TransitionClassifier) TransitionWithConf(c Transition.Configuration) (Transition.Configuration, Transition.Transition) {
	var (
		bestScore             float64
		bestConf, currentConf Transition.Configuration
		bestTransition        Transition.Transition
	)
	tChan := tc.TransFunc.YieldTransitions(c)
	for transition := range tChan {
		currentConf = tc.TransFunc.Transition(c, transition)
		currentScore := tc.ScoreWithConf(currentConf)
		if tc.ShowConsiderations {
			fmt.Println("\t\tConsidering transition", transition, "\t", currentScore)
		}
		if bestConf == nil || currentScore > bestScore {
			bestScore, bestConf, bestTransition = currentScore, currentConf, transition
		}
	}
	if bestConf == nil {
		panic("Got no best transition - what's going on here?")
	}
	if tc.ShowConsiderations {
		fmt.Println("\tChose transition", bestTransition)
	}
	return bestConf, bestTransition
}

type PerceptronModel struct {
	PerceptronModel *Perceptron.LinearPerceptron
}

var _ Dependency.ParameterModel = &PerceptronModel{}

func (p *PerceptronModel) WeightedValue(val Dependency.ParameterModelValue) Dependency.ParameterModelValue {
	vec := val.(*PerceptronModelValue).vector
	return Dependency.ParameterModelValue(&PerceptronModelValue{p.PerceptronModel.Weights.Weighted(vec)})
}

func (p *PerceptronModel) NewModelValue() Dependency.ParameterModelValue {
	newVector := make(Perceptron.SparseWeightVector)
	return Dependency.ParameterModelValue(&PerceptronModelValue{&newVector})
}

func (p *PerceptronModel) ModelValue(val interface{}) Dependency.ParameterModelValue {
	features := val.([]Perceptron.Feature)
	featuresAsWeights := p.PerceptronModel.Weights.FeatureWeights(features)
	return Dependency.ParameterModelValue(&PerceptronModelValue{featuresAsWeights})
}

func (p *PerceptronModel) ModelValueOnes(val interface{}) Dependency.ParameterModelValue {
	features := val.([]Perceptron.Feature)
	featuresAsWeights := Perceptron.NewVectorOfOnesFromFeatures(features)
	return Dependency.ParameterModelValue(&PerceptronModelValue{featuresAsWeights})
}

func (p *PerceptronModel) Model() interface{} {
	return p.PerceptronModel
}

type PerceptronModelValue struct {
	vector *Perceptron.SparseWeightVector
}

var _ Dependency.ParameterModelValue = &PerceptronModelValue{}

func (pmv *PerceptronModelValue) Score() float64 {
	return pmv.vector.L1Norm()
}

func (pmv *PerceptronModelValue) ValueWith(other interface{}) Dependency.ParameterModelValue {
	otherVec := other.(*PerceptronModelValue)
	return Dependency.ParameterModelValue(&PerceptronModelValue{pmv.vector.Add(otherVec.vector)})
}

func (pmv *PerceptronModelValue) Increment(other interface{}) {
	featureVec := other.(*PerceptronModelValue)
	pmv.vector.UpdateAdd(featureVec.vector)
}

func (pmv *PerceptronModelValue) Decrement(other interface{}) {
	featureVec := other.(*Perceptron.SparseWeightVector)
	pmv.vector.UpdateSubtract(featureVec)
}

func (pmv *PerceptronModelValue) Copy() Dependency.ParameterModelValue {
	return Dependency.ParameterModelValue(&PerceptronModelValue{pmv.vector.Copy()})
}

func ArrayDiff(left []Perceptron.Feature, right []Perceptron.Feature) ([]string, []string) {
	var (
		leftStr, rightStr   []string = make([]string, len(left)), make([]string, len(right))
		onlyLeft, onlyRight []string = make([]string, 0, len(left)), make([]string, 0, len(right))
	)
	for i, val := range left {
		leftStr[i] = string(val)
	}
	for i, val := range right {
		rightStr[i] = string(val)
	}
	sort.Strings(leftStr)
	sort.Strings(rightStr)
	i, j := 0, 0
	for i < len(leftStr) || j < len(rightStr) {
		switch {
		case i < len(leftStr) && j < len(rightStr):
			comp := Util.Strcmp(leftStr[i], rightStr[j])
			switch {
			case comp == 0:
				i++
				j++
			case comp < 0:
				onlyLeft = append(onlyLeft, leftStr[i])
				i++
			case comp > 0:
				onlyRight = append(onlyRight, rightStr[j])
				j++
			}
		case i < len(leftStr):
			onlyLeft = append(onlyLeft, leftStr[i])
			i++
		case j < len(rightStr):
			onlyRight = append(onlyRight, rightStr[j])
			j++
		}
	}
	return onlyLeft, onlyRight
}
