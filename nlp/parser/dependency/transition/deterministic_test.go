package Transition

import (
	"chukuparser/algorithm/featurevector"

	"chukuparser/algorithm/perceptron"
	"chukuparser/algorithm/transition"
	TransitionModel "chukuparser/algorithm/transition/model"
	"chukuparser/nlp/parser/dependency"
	"chukuparser/util"
	"log"
	"runtime"
	"sort"
	"testing"
)

func TestDeterministic(t *testing.T) {
	SetupTestEnum()
	SetupEagerTransEnum()
	runtime.GOMAXPROCS(runtime.NumCPU())
	extractor := &GenericExtractor{
		EFeatures: Util.NewEnumSet(len(TEST_RICH_FEATURES)),
	}
	extractor.Init()
	// verify load
	for _, featurePair := range TEST_RICH_FEATURES {
		if err := extractor.LoadFeature(featurePair[0], featurePair[1]); err != nil {
			t.Error("Failed to load feature", err.Error())
			t.FailNow()
		}
	}

	arcSystem := &ArcEager{
		ArcStandard: ArcStandard{
			SHIFT:       SH,
			LEFT:        LA,
			RIGHT:       RA,
			Relations:   TEST_ENUM_RELATIONS,
			Transitions: TRANSITIONS_ENUM,
		},
		REDUCE:  RE,
		POPROOT: PR,
	}
	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)

	conf := &SimpleConfiguration{
		EWord:  EWord,
		EPOS:   EPOS,
		EWPOS:  EWPOS,
		ERel:   TEST_ENUM_RELATIONS,
		ETrans: TRANSITIONS_ENUM,
	}

	deterministic := &Deterministic{
		TransFunc:          transitionSystem,
		FeatExtractor:      extractor,
		ReturnModelValue:   true,
		ReturnSequence:     true,
		ShowConsiderations: false,
		Base:               conf,
		NoRecover:          true,
	}
	decoder := Perceptron.EarlyUpdateInstanceDecoder(deterministic)
	updater := new(TransitionModel.AveragedModelStrategy)

	model := TransitionModel.NewAvgMatrixSparse(extractor.EFeatures.Len(), nil)
	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init(model)
	goldModel := Dependency.TransitionParameterModel(&PerceptronModel{model})

	_, goldParams := deterministic.ParseOracle(GetTestDepGraph(), nil, goldModel)
	if goldParams == nil {
		t.Fatal("Got nil params from deterministic oracle parsing, can't test deterministic-perceptron model")
	}
	seq := goldParams.(*ParseResultParameters).Sequence

	goldSequence := make(ScoredConfigurations, len(seq))
	var (
		lastFeatures *Transition.FeaturesList
		curFeats     []FeatureVector.Feature
	)
	for i := len(seq) - 1; i >= 0; i-- {
		val := seq[i]
		curFeats = extractor.Features(val)
		lastFeatures = &Transition.FeaturesList{curFeats, val.GetLastTransition(), lastFeatures}
		goldSequence[len(seq)-i-1] = &ScoredConfiguration{val.(DependencyConfiguration), val.GetLastTransition(), 0.0, lastFeatures, 0, 0, true}
	}

	goldInstances := []Perceptron.DecodedInstance{
		&Perceptron.Decoded{Perceptron.Instance(rawTestSent), goldSequence}}
	// log.Println(goldSequence)
	// train with increasing iterations
	convergenceIterations := []int{1, 8, 16, 32}
	// convergenceIterations := []int{4}
	convergenceSharedSequence := make([]int, 0, len(convergenceIterations))
	for _, iterations := range convergenceIterations {
		perceptron.Iterations = iterations
		// perceptron.Log = true
		model = TransitionModel.NewAvgMatrixSparse(extractor.EFeatures.Len(), nil)
		perceptron.Init(model)

		// deterministic.ShowConsiderations = true
		perceptron.Train(goldInstances)

		parseModel := Dependency.TransitionParameterModel(&PerceptronModel{model})
		deterministic.ShowConsiderations = false
		_, params := deterministic.Parse(TEST_SENT, nil, parseModel)
		seq := params.(*ParseResultParameters).Sequence
		sharedSteps := goldSequence[len(goldSequence)-1].C.Conf().GetSequence().SharedTransitions(seq)
		convergenceSharedSequence = append(convergenceSharedSequence, sharedSteps)
	}

	// verify convergence
	log.Println(convergenceSharedSequence)
	if !sort.IntsAreSorted(convergenceSharedSequence) || convergenceSharedSequence[0] == convergenceSharedSequence[len(convergenceSharedSequence)-1] {
		t.Error("Model not converging, shared sequences lengths:", convergenceSharedSequence)
	}
}

func TestArrayDiff(t *testing.T) {
	left := []FeatureVector.Feature{"def", "abc"}
	right := []FeatureVector.Feature{"def", "ghi"}
	oLeft, oRight := ArrayDiff(left, right)
	if len(oLeft) != 1 {
		t.Error("Wrong len for oLeft", oLeft)
	}
	if len(oRight) != 1 {
		t.Error("Wrong len for oRight", oRight)
	}
	if len(oLeft) > 0 && oLeft[0] != "abc" {
		t.Error("Didn't get abc for oLeft")
	}
	if len(oRight) > 0 && oRight[0] != "ghi" {
		t.Error("Didn't get ghi for oRight")
	}
}
