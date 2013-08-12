package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP/Parser/Dependency"
	"runtime"
	"sort"
	"testing"
)

func TestBeam(t *testing.T) {
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

	beam := &Beam{Transition: transitionSystem, FeatExtractor: extractor}
	decoder := Perceptron.EarlyUpdateInstanceDecoder(beam)
	updater := new(Perceptron.AveragedStrategy)

	goldInstances := []Perceptron.DecodedInstance{
		&Perceptron.Decoded{Perceptron.Instance(TEST_SENT), GetTestDepGraph()}}

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}

	// get gold parse
	goldModel := Dependency.ParameterModel(&PerceptronModel{perceptron})
	deterministic := &Deterministic{transitionSystem, extractor, true, true, false}
	_, goldParams := deterministic.ParseOracle(TEST_SENT, GetTestDepGraph(), nil, goldModel)
	goldSequence := goldParams.(*ParseResultParameters).sequence

	// train with increasing iterations
	convergenceIterations := []int{1, 5, 10}
	convergenceSharedSequence := make([]int, 0, len(convergenceIterations))
	for _, iterations := range convergenceIterations {
		perceptron.Iterations = iterations
		perceptron.Init()

		perceptron.Train(goldInstances)

		model := Dependency.ParameterModel(&PerceptronModel{perceptron})
		_, params := beam.Parse(TEST_SENT, nil, model)
		sharedSteps := 0
		if params != nil {
			seq := params.(*ParseResultParameters).sequence
			sharedSteps = goldSequence.SharedTransitions(seq)
		}
		convergenceSharedSequence = append(convergenceSharedSequence, sharedSteps)
	}
	if len(convergenceSharedSequence) != len(convergenceIterations) {
		t.Error("Not enough examples in shared sequence samples")
	}
	// verify convergence
	if !sort.IntsAreSorted(convergenceSharedSequence) || convergenceSharedSequence[len(convergenceSharedSequence)-1] == 0 {
		t.Error("Model not converging, shared sequences lengths:", convergenceSharedSequence)
	}
}
