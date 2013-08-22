package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP/Parser/Dependency"
	// "log"
	"runtime"
	"sort"
	"testing"
)

func TestDeterministic(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	extractor := new(GenericExtractor)
	// verify load
	for _, feature := range TEST_RICH_FEATURES {
		if err := extractor.LoadFeature(feature); err != nil {
			t.Error("Failed to load feature", err.Error())
			t.FailNow()
		}
	}
	arcSystem := &ArcEager{}
	arcSystem.Relations = TEST_RELATIONS
	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)

	deterministic := &Deterministic{transitionSystem, extractor, true, true, false, &SimpleConfiguration{}}
	decoder := Perceptron.EarlyUpdateInstanceDecoder(deterministic)
	updater := new(Perceptron.AveragedStrategy)

	goldInstances := []Perceptron.DecodedInstance{
		&Perceptron.Decoded{Perceptron.Instance(TEST_SENT), GetTestDepGraph()}}

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init()
	goldModel := Dependency.ParameterModel(&PerceptronModel{perceptron})

	_, goldParams := deterministic.ParseOracle(GetTestDepGraph(), nil, goldModel)
	goldSequence := goldParams.(*ParseResultParameters).Sequence
	// log.Println(goldSequence.String())

	// train with increasing iterations
	convergenceIterations := []int{1, 8, 16, 32}
	convergenceSharedSequence := make([]int, 0, len(convergenceIterations))
	for _, iterations := range convergenceIterations {
		perceptron.Iterations = iterations
		// perceptron.Log = true
		perceptron.Init()

		// deterministic.ShowConsiderations = true
		perceptron.Train(goldInstances)

		model := Dependency.ParameterModel(&PerceptronModel{perceptron})
		// deterministic.ShowConsiderations = true
		_, params := deterministic.Parse(TEST_SENT, nil, model)
		seq := params.(*ParseResultParameters).Sequence
		sharedSteps := goldSequence.SharedTransitions(seq)
		convergenceSharedSequence = append(convergenceSharedSequence, sharedSteps)
	}

	// verify convergence
	if !sort.IntsAreSorted(convergenceSharedSequence) || convergenceSharedSequence[0] == convergenceSharedSequence[len(convergenceSharedSequence)-1] {
		t.Error("Model not converging, shared sequences lengths:", convergenceSharedSequence)
	}
}

func TestArrayDiff(t *testing.T) {
	left := []Perceptron.Feature{"def", "abc"}
	right := []Perceptron.Feature{"def", "ghi"}
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
