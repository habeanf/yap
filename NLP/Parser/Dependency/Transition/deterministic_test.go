package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP/Parser/Dependency"
	"runtime"
	"sort"
	"testing"
)

//ALL RICH FEATURES
var TEST_DEP_FEATURES []string = []string{
	"S0|w|p",
	"S0|w",
	"S0|p",
	"N0|w|p",
	"N0|w",
	"N0|p",
	"N1|w|p",
	"N1|w",
	"N1|p",
	"N2|w|p",
	"N2|w",
	"N2|p",
	"S0|w|p+N0|w|p",
	"S0|w|p+N0|w",
	"S0|w+N0|w|p",
	"S0|w|p+N0|p",
	"S0|p+N0|w|p",
	"S0|w+N0|w",
	"S0|p+N0|p",
	"N0|p+N1|p",
	"N0|p+N1|p+N2|p",
	"S0|p+N0|p+N1|p",
	"S0h|p+S0|p+N0|p",
	"S0|p+S0l|p+N0|p",
	"S0|p+S0r|p+N0|p",
	"S0|p+N0|p+N0l|p",
	"S0|w|d",
	"S0|p|d",
	"N0|w|d",
	"N0|p|d",
	"S0|w+N0|w|d",
	"S0|p+N0|p|d",
	"S0|w|vr",
	"S0|p|vr",
	"S0|w|vl",
	"S0|p|vl",
	"N0|w|vl",
	"N0|p|vl",
	"S0h|w",
	"S0h|p",
	"S0|l",
	"S0l|w",
	"S0l|p",
	"S0l|l",
	"S0r|w",
	"S0r|p",
	"S0r|l",
	"N0l|w",
	"N0l|p",
	"N0l|l",
	"S0h2|w",
	"S0h2|p",
	"S0h|l",
	"S0l2|w",
	"S0l2|p",
	"S0l2|l",
	"S0r2|w",
	"S0r2|p",
	"S0r2|l",
	"N0l2|w",
	"N0l2|p",
	"N0l2|l",
	"S0|p+S0l|p+S0l2|p",
	"S0|p+S0r|p+S0r2|p",
	"S0|p+S0h|p+S0h2|p",
	"N0|p+N0l|p+N0l2|p",
	"S0|w|sr",
	"S0|p|sr",
	"S0|w|sl",
	"S0|p|sl",
	"N0|w|sl",
	"N0|p|sl"}

func TestDeterministic(t *testing.T) {
	runtime.GOMAXPROCS(4)
	extractor := new(GenericExtractor)
	// verify load
	for _, feature := range TEST_DEP_FEATURES {
		if err := extractor.LoadFeature(feature); err != nil {
			t.Error("Failed to load feature", err.Error())
			t.FailNow()
		}
	}
	arcSystem := &ArcEager{}
	arcSystem.Relations = TEST_RELATIONS
	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)
	deterministic := &Deterministic{transitionSystem, extractor, true, true, false}
	decoder := Perceptron.EarlyUpdateInstanceDecoder(deterministic)
	updater := new(Perceptron.AveragedStrategy)

	goldInstances := []Perceptron.DecodedInstance{
		&Perceptron.Decoded{Perceptron.Instance(TEST_SENT), GetTestDepGraph()}}

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	goldModel := Dependency.ParameterModel(&PerceptronModel{perceptron})

	_, goldParams := deterministic.ParseOracle(TEST_SENT, GetTestDepGraph(), nil, goldModel)
	goldSequence := goldParams.(*ParseResultParameters).sequence

	// train with increasing iterations
	convergenceIterations := []int{1, 5, 10}
	convergenceSharedSequence := make([]int, 0, len(convergenceIterations))
	for _, iterations := range convergenceIterations {
		perceptron.Iterations = iterations
		perceptron.Init()

		deterministic.ShowConsiderations = false
		perceptron.Train(goldInstances)

		model := Dependency.ParameterModel(&PerceptronModel{perceptron})
		deterministic.ShowConsiderations = false
		_, params := deterministic.Parse(TEST_SENT, nil, model)
		seq := params.(*ParseResultParameters).sequence
		sharedSteps := goldSequence.SharedTransitions(seq)
		convergenceSharedSequence = append(convergenceSharedSequence, sharedSteps)
	}

	// verify convergence
	if !sort.IntsAreSorted(convergenceSharedSequence) {
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
