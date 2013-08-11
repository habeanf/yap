package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP/Parser/Dependency"
	"fmt"
	"log"
	"runtime"
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
	var iterations int = 10
	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater, Iterations: iterations}
	perceptron.Init()

	goldInstance := Perceptron.Decoded{Perceptron.Instance(TEST_SENT), GetTestDepGraph()}
	goldInstanceChan := make(chan Perceptron.DecodedInstance, 1)
	goldInstanceChan <- Perceptron.DecodedInstance(&goldInstance)
	close(goldInstanceChan)
	deterministic.ShowConsiderations = true
	perceptron.Train(goldInstanceChan)
	log.Println("Trained perceptron")

	model := Dependency.ParameterModel(&PerceptronModel{perceptron})
	deterministic.ShowConsiderations = false
	graph, params := deterministic.Parse(TEST_SENT, nil, model)
	if !graph.Equal(GetTestDepGraph()) {
		t.Error("Parsed graph does not equal test graph")
	}
	fmt.Println(params.(*ParseResultParameters).sequence.String())
}

func BenchmarkDeterministic(b *testing.B) {
	runtime.GOMAXPROCS(4)
	extractor := new(GenericExtractor)
	// verify load
	for _, feature := range TEST_DEP_FEATURES {
		if err := extractor.LoadFeature(feature); err != nil {
			b.Error("Failed to load feature", err.Error())
			b.FailNow()
		}
	}
	arcSystem := &ArcEager{}
	arcSystem.Relations = TEST_RELATIONS
	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)
	deterministic := &Deterministic{transitionSystem, extractor, true, true, false}
	decoder := Perceptron.EarlyUpdateInstanceDecoder(deterministic)
	updater := new(Perceptron.AveragedStrategy)

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater, Iterations: 2}
	perceptron.Init()

	goldInstance := Perceptron.Decoded{Perceptron.Instance(TEST_SENT), GetTestDepGraph()}
	goldInstanceChan := make(chan Perceptron.DecodedInstance, 1)
	goldInstanceChan <- Perceptron.DecodedInstance(&goldInstance)
	close(goldInstanceChan)
	perceptron.Train(goldInstanceChan)
	fmt.Println("Trained perceptron")

	model := Dependency.ParameterModel(&PerceptronModel{perceptron})
	graph, _ := deterministic.Parse(TEST_SENT, nil, model)
	if !graph.Equal(GetTestDepGraph()) {
		b.Error("Parsed graph does not equal test graph")
	}

}
