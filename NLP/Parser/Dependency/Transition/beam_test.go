package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP/Parser/Dependency"
	"chukuparser/Util"
	"log"
	"runtime"
	"sort"
	"testing"
)

func TestBeam(t *testing.T) {
	SetupEagerTransEnum()
	SetupTestEnum()
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	// runtime.GOMAXPROCS(1)
	runtime.GOMAXPROCS(runtime.NumCPU())
	extractor := &GenericExtractor{
		EFeatures: Util.NewEnumSet(len(TEST_RICH_FEATURES)),
	}
	// verify load
	for _, feature := range TEST_RICH_FEATURES {
		if err := extractor.LoadFeature(feature); err != nil {
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
		REDUCE: RE,
	}
	arcSystem.Relations = TEST_ENUM_RELATIONS
	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)
	conf := &SimpleConfiguration{
		EWord:  EWord,
		EPOS:   EPOS,
		EWPOS:  EWPOS,
		ERel:   TEST_ENUM_RELATIONS,
		ETrans: TRANSITIONS_ENUM,
	}

	beam := &Beam{
		TransFunc:     transitionSystem,
		FeatExtractor: extractor,
		Base:          conf,
		NumRelations:  arcSystem.Relations.Len(),
	}

	decoder := Perceptron.EarlyUpdateInstanceDecoder(beam)
	updater := new(Perceptron.AveragedStrategy)

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init()

	// get gold parse
	goldModel := Dependency.ParameterModel(&PerceptronModel{perceptron})
	deterministic := &Deterministic{transitionSystem, extractor, true, true, false, conf}
	_, goldParams := deterministic.ParseOracle(GetTestDepGraph(), nil, goldModel)
	if goldParams == nil {
		t.Fatal("Got nil params from deterministic oracle parsing, can't test beam-perceptron model")
	}
	goldSequence := goldParams.(*ParseResultParameters).Sequence

	goldInstances := []Perceptron.DecodedInstance{
		&Perceptron.Decoded{Perceptron.Instance(TEST_SENT), goldSequence[0]}}

	// perceptron.Log = true
	beam.ConcurrentExec = true
	beam.ReturnSequence = true
	// train with increasing iterations
	convergenceIterations := []int{1, 8}
	beamSizes := []int{1, 8}
	for _, beamSize := range beamSizes {
		beam.Size = beamSize
		convergenceSharedSequence := make([]int, 0, len(convergenceIterations))
		for _, iterations := range convergenceIterations {
			perceptron.Iterations = iterations
			perceptron.Init()

			// log.Println("Starting training", iterations, "iterations")
			perceptron.Log = false
			perceptron.Train(goldInstances)
			// log.Println("Finished training", iterations, "iterations")

			model := Dependency.ParameterModel(&PerceptronModel{perceptron})
			beam.ReturnModelValue = false
			_, params := beam.Parse(TEST_SENT, nil, model)
			sharedSteps := 0
			if params != nil {
				seq := params.(*ParseResultParameters).Sequence
				sharedSteps = goldSequence.SharedTransitions(seq)
			}
			convergenceSharedSequence = append(convergenceSharedSequence, sharedSteps)
		}
		if len(convergenceSharedSequence) != len(convergenceIterations) {
			t.Error("Not enough examples in shared sequence samples")
		}
		// verify convergence
		log.Println("Shared Sequence For Beam", beamSize, convergenceSharedSequence)
		if !sort.IntsAreSorted(convergenceSharedSequence) || convergenceSharedSequence[len(convergenceSharedSequence)-1] == 0 {
			t.Error("Model not converging, shared sequences lengths:", convergenceSharedSequence)
		}
	}
	// t.Error("bla")
}
