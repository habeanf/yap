package Transition

import (
	"chukuparser/Algorithm/FeatureVector"
	"chukuparser/Algorithm/Perceptron"
	"chukuparser/Algorithm/Transition"
	TransitionModel "chukuparser/Algorithm/Transition/Model"
	"chukuparser/NLP/Parser/Dependency"
	NLP "chukuparser/NLP/Types"
	"chukuparser/Util"
	"fmt"
	"log"
	"sort"
)

type Deterministic struct {
	TransFunc          Transition.TransitionSystem
	FeatExtractor      Perceptron.FeatureExtractor
	ReturnModelValue   bool
	ReturnSequence     bool
	ShowConsiderations bool
	Base               DependencyConfiguration
	NoRecover          bool
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
	transitionClassifier := &TransitionClassifier{Model: model.(Dependency.TransitionParameterModel), TransFunc: d.TransFunc, FeatExtractor: d.FeatExtractor}
	transitionClassifier.Init()
	transitionClassifier.ShowConsiderations = d.ShowConsiderations

	c := d.Base.Conf().Copy()
	c.(DependencyConfiguration).Clear()
	c.Init(sent)
	var prevConf Transition.Configuration
	// deterministic parsing algorithm
	for !c.Terminal() {
		prevConf = c
		c, _ = transitionClassifier.TransitionWithConf(c)
		if c == nil {
			// log.Println("Got nil configuration!")
			c = prevConf
			break
		}
		transitionClassifier.Increment(c)
	}

	// build result parameters
	var resultParams *ParseResultParameters
	if d.ReturnModelValue || d.ReturnSequence {
		resultParams = new(ParseResultParameters)
		if d.ReturnModelValue {
			resultParams.modelValue = transitionClassifier.FeaturesList
		}
		if d.ReturnSequence {
			resultParams.Sequence = c.GetSequence()
		}
	}
	configurationAsGraph := c.(NLP.DependencyGraph)
	return configurationAsGraph, resultParams
}

func (d *Deterministic) ParseOracle(gold NLP.DependencyGraph, constraints interface{}, model Dependency.ParameterModel) (configurationAsGraph NLP.DependencyGraph, result interface{}) {
	if !d.NoRecover {
		defer func() {
			if r := recover(); r != nil {
				configurationAsGraph = nil
				result = nil
			}
		}()
	}
	if constraints != nil {
		panic("Got non-nil constraints; deterministic dependency parsing does not consider constraints")
	}
	if d.TransFunc == nil {
		panic("Can't parse without a transition system")
	}
	c := d.Base.Conf().Copy()
	c.(DependencyConfiguration).Clear()
	c.Init(gold.Sentence())
	classifier := TransitionClassifier{Model: model.(Dependency.TransitionParameterModel), FeatExtractor: d.FeatExtractor, TransFunc: d.TransFunc}

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
			resultParams.modelValue = classifier.FeaturesList
		}
		if d.ReturnSequence {
			resultParams.Sequence = c.GetSequence()
		}
	}
	configurationAsGraph = c.(NLP.DependencyGraph)
	result = resultParams
	return
}

func (d *Deterministic) ParseOracleEarlyUpdate(sent NLP.Sentence, gold Transition.ConfigurationSequence, constraints interface{}, model Dependency.ParameterModel) (Transition.Configuration, Transition.Configuration, interface{}, interface{}, int) {
	if constraints != nil {
		panic("Got non-nil constraints; deterministic dependency parsing does not consider constraints")
	}
	if d.TransFunc == nil {
		panic("Can't parse without a transition system")
	}

	// Initializations
	c := d.Base.Copy()
	c.(DependencyConfiguration).Clear()
	c.Init(sent)

	classifier := TransitionClassifier{Model: model.(Dependency.TransitionParameterModel), FeatExtractor: d.FeatExtractor, TransFunc: d.TransFunc}
	classifier.ShowConsiderations = d.ShowConsiderations

	classifier.Init()

	var (
		predTrans                          Transition.Transition
		prevConf, goldConf                 Transition.Configuration
		predFeatures                       []FeatureVector.Feature
		goldFeaturesList, predFeaturesList *TransitionModel.FeaturesList
		i                                  int = 0
	)
	prefix := log.Prefix()
	for !c.Terminal() {
		log.SetPrefix(fmt.Sprintf("%s %d ", prefix, i))
		goldConf = gold[i] // Oracle's gold sequence
		// log.Printf("Gold Transition: %s\n", goldConf)
		prevConf = c
		c, predTrans = classifier.TransitionWithConf(c)
		// log.Printf("Pred Transition: %s\n", c)

		// verify the right transition was chosen
		if c == nil || predTrans != goldConf.GetLastTransition() {
			c = prevConf
			// d.FeatExtractor.(*GenericExtractor).Log = true
			predFeatures = d.FeatExtractor.Features(c)
			goldFeatures := d.FeatExtractor.Features(goldConf)
			// d.FeatExtractor.(*GenericExtractor).Log = false
			goldFeaturesList = &TransitionModel.FeaturesList{goldFeatures, goldConf.GetLastTransition(), nil}
			predFeaturesList = &TransitionModel.FeaturesList{predFeatures, predTrans, nil}
			break
		}
		i++
	}

	return c, goldConf, predFeaturesList, goldFeaturesList, i
}

// Perceptron functions
func (d *Deterministic) Decode(instance Perceptron.Instance, m Perceptron.Model) (Perceptron.DecodedInstance, interface{}) {
	sent := instance.(NLP.Sentence)
	transitionModel := m.(TransitionModel.Interface)
	model := Dependency.TransitionParameterModel(&PerceptronModel{transitionModel})
	d.ReturnModelValue = true
	graph, parseParamsInterface := d.Parse(sent, nil, model)
	parseParams := parseParamsInterface.(*ParseResultParameters)
	return &Perceptron.Decoded{instance, graph}, parseParams.modelValue
}

func (d *Deterministic) DecodeGold(goldInstance Perceptron.DecodedInstance, m Perceptron.Model) (Perceptron.DecodedInstance, interface{}) {
	graph := goldInstance.Decoded().(NLP.DependencyGraph)
	transitionModel := m.(TransitionModel.Interface)
	model := Dependency.TransitionParameterModel(&PerceptronModel{transitionModel})
	d.ReturnModelValue = true
	parsedGraph, parseParamsInterface := d.ParseOracle(graph, nil, model)
	if !graph.Equal(parsedGraph) {
		panic("Oracle parse result does not equal gold")
	}
	parseParams := parseParamsInterface.(*ParseResultParameters)
	return &Perceptron.Decoded{goldInstance.Instance(), graph}, parseParams.modelValue
}

func (d *Deterministic) DecodeEarlyUpdate(goldInstance Perceptron.DecodedInstance, m Perceptron.Model) (Perceptron.DecodedInstance, interface{}, interface{}, int) {
	sent := goldInstance.Instance().(NLP.Sentence)

	// abstract casting >:-[
	rawGoldSequence := goldInstance.Decoded().(Transition.Configuration).GetSequence()

	// drop the first (seq are in reverse) configuration, as it is the initial one
	// which is by definition without a score or features
	rawGoldSequence = rawGoldSequence[:len(rawGoldSequence)-1]

	goldSequence := make(Transition.ConfigurationSequence, len(rawGoldSequence))
	for i := len(rawGoldSequence) - 1; i >= 0; i-- {
		goldSequence[len(rawGoldSequence)-i-1] = rawGoldSequence[i]
	}

	transitionModel := m.(TransitionModel.Interface)
	model := Dependency.TransitionParameterModel(&PerceptronModel{transitionModel})
	d.ReturnModelValue = true
	parsedConf, _, parsedWeights, goldWeights, earlyUpdatedAt := d.ParseOracleEarlyUpdate(sent, goldSequence, nil, model)
	return &Perceptron.Decoded{goldInstance.Instance(), parsedConf}, parsedWeights, goldWeights, earlyUpdatedAt
}

type TransitionClassifier struct {
	Model              Dependency.TransitionParameterModel
	TransFunc          Transition.TransitionSystem
	FeatExtractor      Perceptron.FeatureExtractor
	Score              float64
	FeaturesList       *TransitionModel.FeaturesList
	ShowConsiderations bool
}

func (tc *TransitionClassifier) Init() {
	tc.Score = 0.0
}

func (tc *TransitionClassifier) Increment(c Transition.Configuration) *TransitionClassifier {
	features := tc.FeatExtractor.Features(Perceptron.Instance(c))
	tc.FeaturesList = &TransitionModel.FeaturesList{features, c.GetLastTransition(), tc.FeaturesList}
	tc.Score += tc.Model.TransitionModel().TransitionScore(c.GetLastTransition(), features)
	return tc
}

func (tc *TransitionClassifier) ScoreWithConf(c Transition.Configuration) float64 {
	features := tc.FeatExtractor.Features(Perceptron.Instance(c))
	return tc.Score + tc.Model.TransitionModel().TransitionScore(c.GetLastTransition(), features)
}

func (tc *TransitionClassifier) Transition(c Transition.Configuration) Transition.Transition {
	_, transition := tc.TransitionWithConf(c)
	return transition
}

func (tc *TransitionClassifier) TransitionWithConf(c Transition.Configuration) (Transition.Configuration, Transition.Transition) {
	var (
		bestScore, prevScore  float64
		bestConf, currentConf Transition.Configuration
		bestTransition        Transition.Transition
	)
	prevScore = -1
	tChan := tc.TransFunc.YieldTransitions(c)
	for transition := range tChan {
		currentConf = tc.TransFunc.Transition(c, transition)
		currentScore := tc.ScoreWithConf(currentConf)
		if tc.ShowConsiderations && currentScore != prevScore {
			log.Println(" Considering transition", transition, "  ", currentScore, currentConf)
		}
		if bestConf == nil || currentScore > bestScore {
			bestScore, bestConf, bestTransition = currentScore, currentConf, transition
		}
		prevScore = currentScore
	}
	if tc.ShowConsiderations {
		if bestConf != nil {
			log.Println("Chose transition", bestConf.String())
		} else {
			log.Println("No transitions possible")
		}
	}
	return bestConf, bestTransition
}

type PerceptronModel struct {
	PerceptronModel TransitionModel.Interface
}

var _ Dependency.ParameterModel = &PerceptronModel{}

func (p *PerceptronModel) TransitionModel() TransitionModel.Interface {
	return p.PerceptronModel
}

func (p *PerceptronModel) Model() interface{} {
	return p.PerceptronModel
}

type PerceptronModelValue struct {
	vector []FeatureVector.Feature
}

var _ Dependency.ParameterModelValue = &PerceptronModelValue{}

func (pmv *PerceptronModelValue) Clear() {
	pmv.vector = nil
}

func ArrayDiff(left []FeatureVector.Feature, right []FeatureVector.Feature) ([]string, []string) {
	var (
		leftStr, rightStr   []string = make([]string, len(left)), make([]string, len(right))
		onlyLeft, onlyRight []string = make([]string, 0, len(left)), make([]string, 0, len(right))
	)
	for i, val := range left {
		leftStr[i] = val.(string)
	}
	for i, val := range right {
		rightStr[i] = val.(string)
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
