package disambig

import (
	"chukuparser/algorithm/featurevector"
	"chukuparser/algorithm/perceptron"
	"chukuparser/algorithm/transition"
	TransitionModel "chukuparser/algorithm/transition/model"
	"chukuparser/nlp/parser/dependency"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"
	// "fmt"
	"log"
	"sort"
)

var TransEnum *util.EnumSet

type Deterministic struct {
	TransFunc          transition.TransitionSystem
	FeatExtractor      perceptron.FeatureExtractor
	ReturnModelValue   bool
	ReturnSequence     bool
	ShowConsiderations bool
	Base               *MDConfig
	NoRecover          bool
}

var _ perceptron.InstanceDecoder = &Deterministic{}

type ParseResultParameters struct {
	modelValue interface{}
	Sequence   transition.ConfigurationSequence
}

// Parser functions
func (d *Deterministic) ParseOracle(gold *MDConfig, constraints interface{}) (configurationAsMapping nlp.Mappings, result interface{}) {
	if !d.NoRecover {
		defer func() {
			if r := recover(); r != nil {
				configurationAsMapping = nil
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
	c := d.Base.Copy()
	c.(*MDConfig).Clear()
	c.Init(gold.Lattices)

	oracle := d.TransFunc.Oracle()
	oracle.SetGold(gold.Mappings)
	transitionNum := 0
	for !c.Terminal() {
		transition := oracle.Transition(c)
		c = d.TransFunc.Transition(c, transition)
		// log.Println(c)
		transitionNum++
	}

	// build result parameters
	var resultParams *ParseResultParameters
	if d.ReturnModelValue || d.ReturnSequence {
		resultParams = new(ParseResultParameters)
		if d.ReturnModelValue {
			// resultParams.modelValue = classifier.FeaturesList
		}
		if d.ReturnSequence {
			resultParams.Sequence = c.GetSequence()
		}
	}
	// configurationAsGraph = c.(nlp.DependencyGraph)
	result = resultParams
	return
}

func (d *Deterministic) Decode(instance perceptron.Instance, m perceptron.Model) (perceptron.DecodedInstance, interface{}) {
	return nil, nil
}

func (d *Deterministic) DecodeGold(goldInstance perceptron.DecodedInstance, m perceptron.Model) (perceptron.DecodedInstance, interface{}) {
	mappedConfig := goldInstance.Decoded().(*MDConfig)
	d.ReturnModelValue = true
	_, goldParams := d.ParseOracle(mappedConfig, nil)
	// if !graph.Equal(parsedGraph) {
	// if !parsedGraph.Equal(graph) {
	// 	log.Println("Oracle failed for", graph)
	// 	log.Println("Got graph", parsedGraph)
	// 	panic("Oracle parse result does not equal gold")
	// }
	// 	_, goldParams := deterministic.ParseOracle(graph, nil, tempModel)
	if goldParams != nil {
		seq := goldParams.(*ParseResultParameters).Sequence

		goldSequence := make(ScoredConfigurations, len(seq))
		var (
			lastFeatures *transition.FeaturesList
			curFeats     []featurevector.Feature
		)
		for i := len(seq) - 1; i >= 0; i-- {
			val := seq[i]
			curFeats = d.FeatExtractor.Features(val)
			lastFeatures = &transition.FeaturesList{curFeats, val.GetLastTransition(), lastFeatures}
			// log.Println("Gold seq val", i, val)
			goldSequence[len(seq)-i-1] = &ScoredConfiguration{val.(*MDConfig), val.GetLastTransition(), 0.0, lastFeatures, 0, 0, true}
		}

		// log.Println("Gold seq:\n", goldSequence)
		decoded := &perceptron.Decoded{goldInstance.Instance(), goldSequence}
		return decoded, nil
	} else {
		return nil, nil
	}
}

type TransitionClassifier struct {
	Model              dependency.TransitionParameterModel
	TransFunc          transition.TransitionSystem
	FeatExtractor      perceptron.FeatureExtractor
	Score              int64
	FeaturesList       *transition.FeaturesList
	ShowConsiderations bool
}

func (tc *TransitionClassifier) Init() {
	tc.Score = 0.0
}

func (tc *TransitionClassifier) Increment(c transition.Configuration) *TransitionClassifier {
	features := tc.FeatExtractor.Features(perceptron.Instance(c))
	tc.FeaturesList = &transition.FeaturesList{features, c.GetLastTransition(), tc.FeaturesList}
	tc.Score += tc.Model.TransitionModel().TransitionScore(c.GetLastTransition(), features)
	return tc
}

func (tc *TransitionClassifier) ScoreWithConf(c transition.Configuration) int64 {
	features := tc.FeatExtractor.Features(perceptron.Instance(c))
	return tc.Score + tc.Model.TransitionModel().TransitionScore(c.GetLastTransition(), features)
}

func (tc *TransitionClassifier) Transition(c transition.Configuration) transition.Transition {
	_, transition := tc.TransitionWithConf(c)
	return transition
}

func (tc *TransitionClassifier) TransitionWithConf(c transition.Configuration) (transition.Configuration, transition.Transition) {
	var (
		bestScore, prevScore int64
		bestTransition       transition.Transition
		notFirst             bool
	)
	prevScore = -1
	feats := tc.FeatExtractor.Features(c)
	if tc.ShowConsiderations {
		log.Println(" Showing Considerations For", c)
	}
	tChan := tc.TransFunc.YieldTransitions(c)
	for transition := range tChan {
		currentScore := tc.Model.TransitionModel().TransitionScore(transition, feats)
		if tc.ShowConsiderations && currentScore != prevScore {
			log.Println(" Considering transition", transition, "  ", currentScore)
		}
		if !notFirst || currentScore > bestScore {
			bestScore, bestTransition = currentScore, transition
			notFirst = true
		}
		prevScore = currentScore
	}
	if tc.ShowConsiderations {
		if notFirst {
			log.Println("Chose transition", bestTransition)
		} else {
			log.Println("No transitions possible")
		}
	}
	tc.Score += bestScore
	return tc.TransFunc.Transition(c, bestTransition), bestTransition
}

type PerceptronModel struct {
	PerceptronModel TransitionModel.Interface
}

var _ dependency.ParameterModel = &PerceptronModel{}

func (p *PerceptronModel) TransitionModel() TransitionModel.Interface {
	return p.PerceptronModel
}

func (p *PerceptronModel) Model() interface{} {
	return p.PerceptronModel
}

type PerceptronModelValue struct {
	vector []featurevector.Feature
}

var _ dependency.ParameterModelValue = &PerceptronModelValue{}

func (pmv *PerceptronModelValue) Clear() {
	pmv.vector = nil
}

func ArrayDiff(left []featurevector.Feature, right []featurevector.Feature) ([]string, []string) {
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
			comp := util.Strcmp(leftStr[i], rightStr[j])
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
