package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP/Parser/Dependency"
	"log"
	"runtime"
	"sort"
	"testing"
)

func TestBeam(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	// runtime.GOMAXPROCS(1)
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
	conf := &SimpleConfiguration{}

	beam := &Beam{
		TransFunc:     transitionSystem,
		FeatExtractor: extractor,
		Base:          conf,
		NumRelations:  len(arcSystem.Relations),
	}

	decoder := Perceptron.EarlyUpdateInstanceDecoder(beam)
	updater := new(Perceptron.AveragedStrategy)

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init()

	// get gold parse
	goldModel := Dependency.ParameterModel(&PerceptronModel{perceptron})
	deterministic := &Deterministic{transitionSystem, extractor, true, true, false, conf}
	_, goldParams := deterministic.ParseOracle(GetTestDepGraph(), nil, goldModel)
	goldSequence := goldParams.(*ParseResultParameters).Sequence

	goldInstances := []Perceptron.DecodedInstance{
		&Perceptron.Decoded{Perceptron.Instance(TEST_SENT), goldSequence[0]}}

	// perceptron.Log = true
	beam.ConcurrentExec = false
	beam.ReturnSequence = true
	// train with increasing iterations
	convergenceIterations := []int{20}
	beamSizes := []int{32}
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
	log.Println("Time Expanding (pct):\t", beam.DurExpanding.Seconds(), 100*beam.DurExpanding/beam.DurTotal)
	log.Println("Time Inserting (pct):\t", beam.DurInserting.Seconds(), 100*beam.DurInserting/beam.DurTotal)
	log.Println("Time Inserting-Feat (pct):\t", beam.DurInsertFeat.Seconds(), 100*beam.DurInsertFeat/beam.DurTotal)
	log.Println("Time Inserting-Modl (pct):\t", beam.DurInsertModl.Seconds(), 100*beam.DurInsertModl/beam.DurTotal)
	log.Println("Time Inserting-Scrp (pct):\t", beam.DurInsertScrp.Seconds(), 100*beam.DurInsertScrp/beam.DurTotal)
	log.Println("Time Inserting-Scrm (pct):\t", beam.DurInsertScrm.Seconds(), 100*beam.DurInsertScrm/beam.DurTotal)
	log.Println("Time Inserting-Heap (pct):\t", beam.DurInsertHeap.Seconds(), 100*beam.DurInsertHeap/beam.DurTotal)
	log.Println("Time Inserting-Agen (pct):\t", beam.DurInsertAgen.Seconds(), 100*beam.DurInsertAgen/beam.DurTotal)
	log.Println("Time Inserting-Init (pct):\t", beam.DurInsertInit.Seconds(), 100*beam.DurInsertInit/beam.DurTotal)
	log.Println("Total Time:", beam.DurTotal.Seconds())
	t.Error("bla")
}
