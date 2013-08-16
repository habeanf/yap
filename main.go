package main

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
	"chukuparser/NLP/Parser/Dependency"
	. "chukuparser/NLP/Parser/Dependency/Transition"
	"chukuparser/NLP/Util/Conll"
	"chukuparser/NLP/Util/TaggedSentence"
	"chukuparser/Util"

	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	_ "net/http/pprof"
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
		if i%1000 == 0 {
			log.Println("At line", i)
		}
		sent := graph.TaggedSentence()
		_, goldParams := deterministic.ParseOracle(graph, nil, tempModel)
		seq := goldParams.(*ParseResultParameters).Sequence
		decoded := &Perceptron.Decoded{sent, seq[0]}
		instances[i] = decoded
	}
	return instances
}

func ReadTraining(filename string) []Perceptron.DecodedInstance {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	var instances []Perceptron.DecodedInstance
	dec := gob.NewDecoder(file)
	err = dec.Decode(&instances)
	if err != nil {
		panic(err)
	}
	return instances
}

func WriteTraining(instances []Perceptron.DecodedInstance, filename string) {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	enc := gob.NewEncoder(file)
	err = enc.Encode(instances)
	if err != nil {
		panic(err)
	}
}

func Train(trainingSet []Perceptron.DecodedInstance, iterations, beamSize int, features []string, filename string) *Perceptron.LinearPerceptron {
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

	perceptron := &Perceptron.LinearPerceptron{
		Decoder:   decoder,
		Updater:   updater,
		Tempfile:  filename,
		TempLines: 5000}
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
		TransFunc:       transitionSystem,
		FeatExtractor:   extractor,
		Base:            conf,
		Size:            beamSize,
		NumRelations:    len(arcSystem.Relations),
		Model:           model,
		ConcurrentExec:  true,
		ShortTempAgenda: true}

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

func RegisterTypes() {
	gob.Register(Transition.ConfigurationSequence{})
	gob.Register(&BasicDepGraph{})
	gob.Register(&TaggedDepNode{})
	gob.Register(&BasicDepArc{})
	gob.Register(&Beam{})
	gob.Register(&SimpleConfiguration{})
	gob.Register(&ArcEager{})
	gob.Register(&GenericExtractor{})
	gob.Register(&PerceptronModel{})
	gob.Register(&Perceptron.AveragedStrategy{})
	gob.Register(&Perceptron.Decoded{})
	gob.Register(NLP.BasicTaggedSentence{})
	gob.Register(&StackArray{})
	gob.Register(&ArcSetSimple{})
}

func main() {
	trainFile, trainSeqFile := "train.conll", "train.gob"
	inputFile, outputFile := "devi.txt", "devo.txt"
	iterations, beamSize := 1, 64

	modelFile := fmt.Sprintf("model.b%d.i%d", beamSize, iterations)

	var goldSequences []Perceptron.DecodedInstance

	RegisterTypes()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("Configuration")
	log.Println("Train file:\t", trainFile)
	log.Println("Seq file:\t", trainSeqFile)
	log.Println("Input file:\t", inputFile)
	log.Println("Output file:\t", outputFile)
	log.Println("Iterations:\t", iterations)
	log.Println("Beam Size:\t", beamSize)
	log.Println("Model file:\t", modelFile)
	runtime.GOMAXPROCS(runtime.NumCPU())
	// runtime.GOMAXPROCS(1)

	// launch net server for profiling
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	log.Println("Reading training sentences from", trainFile)
	s, e := Conll.ReadFile(trainFile)
	if e != nil {
		log.Println(e)
		return
	}
	log.Println("Read", len(s), "sentences from", trainFile)
	goldGraphs := Conll.Conll2GraphCorpus(s)
	log.Println("Converted from conll to internal format")
	goldSequences = TrainingSequences(goldGraphs, RICH_FEATURES)
	log.Println("Parsing with gold to get training sequences")
	log.Println("Writing training sequences to", trainSeqFile)
	WriteTraining(goldSequences, trainSeqFile)
	log.Println("Loading training sequences from", trainSeqFile)
	goldSequences = ReadTraining(trainSeqFile)
	log.Println("Loaded", len(goldSequences), "training sequences")
	Util.LogMemory()
	log.Println("Training", iterations, "iteration(s)")
	model := Train(goldSequences, iterations, beamSize, RICH_FEATURES, modelFile)
	log.Println("Done Training")
	Util.LogMemory()

	log.Println("Writing final model to", modelFile)
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
