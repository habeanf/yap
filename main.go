package main

import (
	"chukuparser/NLP/Util/Conll"
	"chukuparser/NLP/Util/TaggedSentence"
	"log"

	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
	"chukuparser/NLP/Parser/Dependency"
	. "chukuparser/NLP/Parser/Dependency/Transition"
	"runtime"

	"fmt"
	"os"
)

var (
	RICH_FEATURES []string = []string{
		"S0|w|p", "S0|w", "S0|p", "N0|w|p",
		"N0|w", "N0|p", "N1|w|p", "N1|w",
		"N1|p", "N2|w|p", "N2|w", "N2|p",
		"S0|w|p+N0|w|p", "S0|w|p+N0|w",
		"S0|w+N0|w|p", "S0|w|p+N0|p",
		"S0|p+N0|w|p", "S0|w+N0|w",
		"S0|p+N0|p", "N0|p+N1|p",
		"N0|p+N1|p+N2|p", "S0|p+N0|p+N1|p",
		"S0h|p+S0|p+N0|p", "S0|p+S0l|p+N0|p",
		"S0|p+S0r|p+N0|p", "S0|p+N0|p+N0l|p",
		"S0|w|d", "S0|p|d", "N0|w|d", "N0|p|d",
		"S0|w+N0|w|d", "S0|p+N0|p|d",
		"S0|w|vr", "S0|p|vr", "S0|w|vl", "S0|p|vl", "N0|w|vl", "N0|p|vl",
		"S0h|w", "S0h|p", "S0|l", "S0l|w",
		"S0l|p", "S0l|l", "S0r|w", "S0r|p",
		"S0r|l", "N0l|w", "N0l|p", "N0l|l",
		"S0h2|w", "S0h2|p", "S0h|l", "S0l2|w",
		"S0l2|p", "S0l2|l", "S0r2|w", "S0r2|p",
		"S0r2|l", "N0l2|w", "N0l2|p", "N0l2|l",
		"S0|p+S0l|p+S0l2|p", "S0|p+S0r|p+S0r2|p",
		"S0|p+S0h|p+S0h2|p", "N0|p+N0l|p+N0l2|p",
		"S0|w|sr", "S0|p|sr", "S0|w|sl", "S0|p|sl",
		"N0|w|sl", "N0|p|sl"}

	LABELS []string = []string{
		"AMOD",
		"DEP",
		"NMOD",
		"OBJ",
		"P",
		"PMOD",
		"PRD",
		"ROOT",
		"SBAR",
		"SUB",
		"VC",
		"VMOD",
	}
)

func TrainingSequences(trainingSet []NLP.LabeledDependencyGraph, features []string) []Perceptron.DecodedInstance {
	extractor := new(GenericExtractor)
	// verify feature load
	for _, feature := range features {
		if err := extractor.LoadFeature(feature); err != nil {
			log.Panicln("Failed to load feature", err.Error())
		}
	}
	arcSystem := &ArcEager{}
	arcSystem.Relations = LABELS
	arcSystem.AddDefaultOracle()

	transitionSystem := Transition.TransitionSystem(arcSystem)
	deterministic := &Deterministic{transitionSystem, extractor, true, true, false}

	decoder := Perceptron.EarlyUpdateInstanceDecoder(deterministic)
	updater := new(Perceptron.AveragedStrategy)

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init()
	tempModel := Dependency.ParameterModel(&PerceptronModel{perceptron})

	instances := make([]Perceptron.DecodedInstance, len(trainingSet))
	for i, graph := range trainingSet {
		sent := graph.TaggedSentence()
		_, goldParams := deterministic.ParseOracle(graph, nil, tempModel)
		seq := goldParams.(*ParseResultParameters).Sequence
		decoded := &Perceptron.Decoded{sent, seq[0]}
		instances[i] = decoded
	}
	return instances
}

func Train(trainingSet []Perceptron.DecodedInstance, iterations, beamSize int, features []string) *Perceptron.LinearPerceptron {
	extractor := new(GenericExtractor)
	// verify feature load
	for _, feature := range features {
		if err := extractor.LoadFeature(feature); err != nil {
			log.Panicln("Failed to load feature", err.Error())
		}
	}
	arcSystem := &ArcEager{}
	arcSystem.Relations = LABELS
	arcSystem.AddDefaultOracle()

	transitionSystem := Transition.TransitionSystem(arcSystem)
	conf := DependencyConfiguration(new(SimpleConfiguration))

	beam := &Beam{
		TransFunc:      transitionSystem,
		FeatExtractor:  extractor,
		Base:           conf,
		NumRelations:   len(arcSystem.Relations),
		Size:           beamSize,
		ConcurrentExec: true}
	decoder := Perceptron.EarlyUpdateInstanceDecoder(beam)
	updater := new(Perceptron.AveragedStrategy)

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init()
	perceptron.Log = true

	perceptron.Iterations = iterations

	perceptron.Train(trainingSet)

	return perceptron
}

func Parse(sents []NLP.TaggedSentence, beamSize int, model Dependency.ParameterModel, features []string) []NLP.LabeledDependencyGraph {
	extractor := new(GenericExtractor)
	// verify load
	for _, feature := range features {
		if err := extractor.LoadFeature(feature); err != nil {
			log.Panicln("Failed to load feature", err.Error())
		}
	}
	arcSystem := &ArcEager{}
	arcSystem.Relations = LABELS
	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)

	conf := DependencyConfiguration(new(SimpleConfiguration))

	beam := &Beam{
		TransFunc:      transitionSystem,
		FeatExtractor:  extractor,
		Base:           conf,
		Size:           beamSize,
		NumRelations:   len(arcSystem.Relations),
		Model:          model,
		ConcurrentExec: true}

	parsedGraphs := make([]NLP.LabeledDependencyGraph, len(sents))
	for i, sent := range sents {
		log.Println("Parsing sent", i)
		graph, _ := beam.Parse(sent, nil, model)
		labeled := graph.(NLP.LabeledDependencyGraph)
		parsedGraphs[i] = labeled
	}
	return parsedGraphs
}

func WriteModel(model Perceptron.Model, filename string) {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	model.Write(file)
}

func ReadModel(filename string) *Perceptron.LinearPerceptron {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	model := new(Perceptron.LinearPerceptron)
	model.Read(file)
	return model
}

func main() {
	trainFile := "train.conll"
	inputFile := "devi.txt"
	outputFile := "devo.conll"
	iterations := 20
	beamSize := 64
	modelFile := fmt.Sprintf("model.b%d.i%d", beamSize, iterations)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	runtime.GOMAXPROCS(runtime.NumCPU())

	s, e := Conll.ReadFile(trainFile)
	log.Println("Read", len(s), "sentences from", trainFile)
	goldGraphs := Conll.Conll2GraphCorpus(s)
	log.Println("Converted from conll to internal format")
	if e != nil {
		log.Println(e)
		return
	}
	log.Println("Parsing with gold to get training sequences")
	goldSequences := TrainingSequences(goldGraphs, RICH_FEATURES)
	log.Println("Training")
	model := Train(goldSequences, iterations, beamSize, RICH_FEATURES)

	log.Println("Writing model to", modelFile)
	WriteModel(model, modelFile)
	// model := ReadModel(modelFile)
	// log.Println("Read model from", modelFile)
	sents, e2 := TaggedSentence.ReadFile(inputFile)
	log.Println("Read", len(sents), "from", inputFile)
	if e2 != nil {
		log.Println(e2)
		return
	}

	log.Print("Parsing")
	parsedGraphs := Parse(sents, beamSize, Dependency.ParameterModel(&PerceptronModel{model}), RICH_FEATURES)
	log.Println("Converted to conll")
	graphAsConll := Conll.Graph2ConllCorpus(parsedGraphs)
	log.Println("Wrote", len(graphAsConll), "in conll format to", outputFile)
	Conll.WriteFile(outputFile, graphAsConll)
}
