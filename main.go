package main

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP/Format/Conll"
	"chukuparser/NLP/Format/Lattice"
	"chukuparser/NLP/Format/Segmentation"
	"chukuparser/NLP/Parser/Dependency"
	. "chukuparser/NLP/Parser/Dependency/Transition"
	"chukuparser/NLP/Parser/Dependency/Transition/Morph"
	NLP "chukuparser/NLP/Types"
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
		"N0|w|sl", "N0|p|sl", "N0|t"}

	LABELS []string = []string{
		"advmod", "amod", "appos", "aux",
		"cc", "ccomp", "comp", "complmn",
		"compound", "conj", "cop", "def",
		"dep", "det", "detmod", "gen",
		"ghd", "gobj", "hd", "mod",
		"mwe", "neg", "nn", "null",
		"num", "number", "obj", "parataxis",
		"pcomp", "pobj", "posspmod", "prd",
		"prep", "prepmod", "punct", "qaux",
		"rcmod", "rel", "relcomp", "subj",
		"tmod", "xcomp",
	}
)

func TrainingSequences(trainingSet []*Morph.BasicMorphGraph, features []string) []Perceptron.DecodedInstance {
	extractor := new(GenericExtractor)
	// verify feature load
	for _, feature := range features {
		if err := extractor.LoadFeature(feature); err != nil {
			log.Panicln("Failed to load feature", err.Error())
		}
	}
	arcSystem := &Morph.ArcEagerMorph{}
	arcSystem.Relations = LABELS
	arcSystem.AddDefaultOracle()

	idleSystem := &Morph.Idle{arcSystem}
	transitionSystem := Transition.TransitionSystem(idleSystem)
	deterministic := &Deterministic{transitionSystem, extractor, false, true, false, &Morph.MorphConfiguration{}}

	decoder := Perceptron.EarlyUpdateInstanceDecoder(deterministic)
	updater := new(Perceptron.AveragedStrategy)

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init()
	tempModel := Dependency.ParameterModel(&PerceptronModel{perceptron})

	instances := make([]Perceptron.DecodedInstance, 0, len(trainingSet))
	for i, graph := range trainingSet {
		if i%100 == 0 {
			log.Println("At line", i)
		}
		sent := graph.Lattice
		// log.Println("Gold parsing graph (nodes, arcs, lattice)")
		// log.Println("Nodes:")
		// for _, node := range graph.Nodes {
		// 	log.Println("\t", node)
		// }
		// log.Println("Arcs:")
		// for _, arc := range graph.Arcs {
		// 	log.Println("\t", arc)
		// }
		// log.Println("Mappings:")
		// for _, m := range graph.Mappings {
		// 	log.Println("\t", m)
		// }
		// log.Println("Lattices:")
		// for _, lat := range graph.Lattice {
		// 	log.Println("\t", lat)
		// }
		_, goldParams := deterministic.ParseOracle(graph, nil, tempModel)
		if goldParams != nil {
			seq := goldParams.(*ParseResultParameters).Sequence
			// log.Println("Gold seq:\n", seq)
			decoded := &Perceptron.Decoded{sent, seq[0]}
			instances = append(instances, decoded)
		}
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
	arcSystem := &Morph.ArcEagerMorph{}
	arcSystem.Relations = LABELS
	arcSystem.AddDefaultOracle()

	idleSystem := &Morph.Idle{arcSystem}
	transitionSystem := Transition.TransitionSystem(idleSystem)
	conf := &Morph.MorphConfiguration{}

	beam := Beam{
		TransFunc:      transitionSystem,
		FeatExtractor:  extractor,
		Base:           conf,
		NumRelations:   len(arcSystem.Relations),
		Size:           beamSize,
		ConcurrentExec: true}
	varbeam := &VarBeam{beam}
	decoder := Perceptron.EarlyUpdateInstanceDecoder(varbeam)
	updater := new(Perceptron.AveragedStrategy)

	perceptron := &Perceptron.LinearPerceptron{
		Decoder:   decoder,
		Updater:   updater,
		Tempfile:  filename,
		TempLines: 1000}

	perceptron.Iterations = iterations
	perceptron.Init()
	// perceptron.TempLoad("model.b64.i1")
	perceptron.Log = true

	perceptron.Train(trainingSet)

	return perceptron
}

func Parse(sents []NLP.LatticeSentence, beamSize int, model Dependency.ParameterModel, features []string) []NLP.MorphDependencyGraph {
	extractor := new(GenericExtractor)
	// verify load
	for _, feature := range features {
		if err := extractor.LoadFeature(feature); err != nil {
			log.Panicln("Failed to load feature", err.Error())
		}
	}
	arcSystem := &Morph.ArcEagerMorph{}
	arcSystem.Relations = LABELS
	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)

	conf := &Morph.MorphConfiguration{}

	beam := Beam{
		TransFunc:       transitionSystem,
		FeatExtractor:   extractor,
		Base:            conf,
		Size:            beamSize,
		NumRelations:    len(arcSystem.Relations),
		Model:           model,
		ConcurrentExec:  true,
		ShortTempAgenda: true}

	varbeam := &VarBeam{beam}

	parsedGraphs := make([]NLP.MorphDependencyGraph, len(sents))
	for i, sent := range sents {
		log.Println("Parsing sent", i)
		graph, _ := varbeam.Parse(sent, nil, model)
		labeled := graph.(NLP.MorphDependencyGraph)
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
	gob.Register(&Morph.BasicMorphGraph{})
	gob.Register(&NLP.Morpheme{})
	gob.Register(&BasicDepArc{})
	gob.Register(&Beam{})
	gob.Register(&Morph.MorphConfiguration{})
	gob.Register(&Morph.ArcEagerMorph{})
	gob.Register(&GenericExtractor{})
	gob.Register(&PerceptronModel{})
	gob.Register(&Perceptron.AveragedStrategy{})
	gob.Register(&Perceptron.Decoded{})
	gob.Register(NLP.LatticeSentence{})
	gob.Register(&StackArray{})
	gob.Register(&ArcSetSimple{})
}

func oldmain() {
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
	log.Println("Converting from conll to internal format")
	// goldGraphs := Conll.Conll2GraphCorpus(s)
	var goldGraphs []*Morph.BasicMorphGraph
	log.Println("Parsing with gold to get training sequences")
	goldSequences = TrainingSequences(goldGraphs, RICH_FEATURES)
	// log.Println("Writing training sequences to", trainSeqFile)
	// WriteTraining(goldSequences, trainSeqFile)
	// log.Println("Loading training sequences from", trainSeqFile)
	// goldSequences = ReadTraining(trainSeqFile)
	log.Println("Successfully Loaded", len(goldSequences), "training sequences")
	log.Println("Running GC")
	log.Println("Before")
	Util.LogMemory()
	runtime.GC()
	log.Println("After")
	Util.LogMemory()
	log.Println("Training", iterations, "iteration(s)")
	model := Train(goldSequences, iterations, beamSize, RICH_FEATURES, modelFile)
	log.Println("Done Training")
	Util.LogMemory()

	log.Println("Writing final model to", modelFile)
	WriteModel(model, modelFile)
	// model := ReadModel(modelFile)
	// log.Println("Read model from", modelFile)
	// sents, e2 := TaggedSentence.ReadFile(inputFile)
	// log.Println("Read", len(sents), "from", inputFile)
	// if e2 != nil {
	// 	log.Println(e2)
	// 	return
	// }
	var sents []NLP.LatticeSentence
	log.Print("Parsing")
	parsedGraphs := Parse(sents, beamSize, Dependency.ParameterModel(&PerceptronModel{model}), RICH_FEATURES)
	log.Println("Converting to conll")
	// graphAsConll := Conll.Graph2ConllCorpus(parsedGraphs)
	log.Println("Wrote", len(parsedGraphs), "in conll format to", outputFile)
	// Conll.WriteFile(outputFile, graphAsConll)
}

func CombineTrainingInputs(graphs []NLP.LabeledDependencyGraph, goldLats, ambLats []NLP.LatticeSentence) ([]*Morph.BasicMorphGraph, int) {
	if len(graphs) != len(goldLats) || len(graphs) != len(ambLats) {
		panic(fmt.Sprintf("Got mismatched training slice inputs (graphs, gold lattices, ambiguous lattices):", len(graphs), len(goldLats), len(ambLats)))
	}
	morphGraphs := make([]*Morph.BasicMorphGraph, len(graphs))
	var (
		numLatticeNoGold int
		noGold           bool
	)
	prefix := log.Prefix()
	for i, goldGraph := range graphs {
		goldLat := goldLats[i]
		ambLat := ambLats[i]
		log.SetPrefix(fmt.Sprintf("%v graph# %v ", prefix, i))
		morphGraphs[i], noGold = Morph.CombineToGoldMorph(goldGraph, goldLat, ambLat)
		if noGold {
			numLatticeNoGold++
		}
	}
	log.SetPrefix(prefix)
	return morphGraphs, numLatticeNoGold
}

func main() {
	// trainFileConll := "train4k.hebtb.gold.conll"
	// trainFileLat := "train4k.hebtb.gold.lattices"
	// trainFileLatPred := "train4k.hebtb.pred.lattices"
	// inputLatPred := "dev.hebtb.pred.conll.tobeparsed.pred_tagged+pred_token.nodisamb.lattices"
	// outputFile := "dev.hebtb.pred.conll"
	// segFile := "dev.hebtb.pred.segmentation"
	// goldSegFile := "train4k.hebtb.gold.segmentation"
	trainFileConll := "dev.hebtb.gold.conll"
	trainFileLat := "dev.hebtb.gold.conll.tobeparsed.gold_tagged+gold_fixed_token.lattices"
	trainFileLatPred := "dev.hebtb.pred.conll.tobeparsed.pred_tagged+pred_token.nodisamb.lattices"
	inputLatPred := "dev.hebtb.pred.conll.tobeparsed.pred_tagged+pred_token.nodisamb.lattices"
	outputFile := "dev.hebtb.pred.conll"
	segFile := "dev.hebtb.pred.segmentation"
	goldSegFile := "dev.hebtb.gold.segmentation"
	// trainFileConll := "dev.hebtb.1.gold.conll"
	// trainFileLat := "dev.hebtb.1.gold.conll.tobeparsed.gold_tagged+gold_fixed_token.lattices"
	// trainFileLatPred := "dev.hebtb.1.pred.conll.tobeparsed.pred_tagged+pred_token.nodisamb.lattices"
	// inputLatPred := "dev.hebtb.1.pred.conll.tobeparsed.pred_tagged+pred_token.nodisamb.lattices"
	// outputFile := "dev.hebtb.1.pred.conll"
	// segFile := "dev.hebtb.1.pred.segmentation"
	// goldSegFile := "dev.hebtb.1.gold.segmentation"

	iterations, beamSize := 1, 4

	modelFile := fmt.Sprintf("model.morph.b%d.i%d", beamSize, iterations)

	// var goldSequences []Perceptron.DecodedInstance

	RegisterTypes()
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("Configuration")
	log.Println("IDLE + All CPOS tags of queue + Agreement")
	log.Println("CPUs:", runtime.NumCPU())
	log.Println("Train file (conll):\t\t", trainFileConll)
	log.Println("Train file (lattice disamb.):\t", trainFileLat)
	log.Println("Train file (lattice ambig.):\t", trainFileLatPred)
	log.Println("Test file (lattice ambig.):\t", inputLatPred)
	// log.Println("Output file:\t", outputFile)
	log.Println("Iterations:\t", iterations)
	log.Println("Beam Size:\t", beamSize)
	log.Println("Model file:\t", modelFile)
	// runtime.GOMAXPROCS(1)

	// launch net server for profiling
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	log.Println("Reading training conll sentences from", trainFileConll)
	s, e := Conll.ReadFile(trainFileConll)
	if e != nil {
		log.Println(e)
		return
	}
	log.Println("Read", len(s), "sentences from", trainFileConll)
	log.Println("Converting from conll to internal structure")
	goldConll := Conll.Conll2GraphCorpus(s)

	log.Println("Reading training disambiguated lattices from", trainFileLat)
	lDis, lDisE := Lattice.ReadFile(trainFileLat)
	if lDisE != nil {
		log.Println(lDisE)
		return
	}
	log.Println("Read", len(lDis), "disambiguated lattices from", trainFileLat)
	log.Println("Converting lattice format to internal structure")
	goldDisLat := Lattice.Lattice2SentenceCorpus(lDis)

	log.Println("Reading ambiguous lattices from", inputLatPred)
	lAmb, lAmbE := Lattice.ReadFile(trainFileLatPred)
	if lAmbE != nil {
		log.Println(lAmbE)
		return
	}
	log.Println("Read", len(lAmb), "ambiguous lattices from", inputLatPred)
	log.Println("Converting lattice format to internal structure")
	goldAmbLat := Lattice.Lattice2SentenceCorpus(lAmb)

	log.Println("Combining into a single gold morph graph with lattices")
	combined, missingGold := CombineTrainingInputs(goldConll, goldDisLat, goldAmbLat)

	log.Println("Combined", len(combined), "graphs, with", missingGold, "missing at least one gold path in lattice")

	log.Println("Parsing with gold to get training sequences")
	goldSequences := TrainingSequences(combined, RICH_FEATURES)
	log.Println("Generated", len(goldSequences), "training sequences")
	// Util.LogMemory()
	log.Println("Training", iterations, "iteration(s)")
	model := Train(goldSequences, iterations, beamSize, RICH_FEATURES, modelFile)
	log.Println("Done Training")
	// Util.LogMemory()

	log.Println("Writing final model to", modelFile)
	WriteModel(model, modelFile)

	// log.Println("Reading model from", modelFile)
	// model := ReadModel(modelFile)
	// log.Println("Read model from", modelFile)
	// sents, e2 := TaggedSentence.ReadFile(inputFile)
	// log.Println("Read", len(sents), "from", inputFile)
	// if e2 != nil {
	// 	log.Println(e2)
	// 	return
	// }

	log.Print("Parsing test")

	log.Println("Reading ambiguous lattices from", inputLatPred)
	lAmb, lAmbE = Lattice.ReadFile(inputLatPred)
	if lAmbE != nil {
		log.Println(lAmbE)
		return
	}

	log.Println("Read", len(lAmb), "ambiguous lattices from", inputLatPred)
	log.Println("Converting lattice format to internal structure")
	predAmbLat := Lattice.Lattice2SentenceCorpus(lAmb)

	parsedGraphs := Parse(predAmbLat, beamSize, Dependency.ParameterModel(&PerceptronModel{model}), RICH_FEATURES)

	log.Println("Converting", len(parsedGraphs), "to conll")
	graphAsConll := Conll.MorphGraph2ConllCorpus(parsedGraphs)
	log.Println("Writing to output file")
	Conll.WriteFile(outputFile, graphAsConll)
	log.Println("Wrote", len(graphAsConll), "in conll format to", outputFile)

	log.Println("Writing to segmentation file")
	Segmentation.WriteFile(segFile, parsedGraphs)
	log.Println("Wrote", len(parsedGraphs), "in segmentation format to", segFile)

	log.Println("Writing to gold segmentation file")
	Segmentation.WriteFile(goldSegFile, ToMorphGraphs(combined))
	log.Println("Wrote", len(combined), "in segmentation format to", goldSegFile)
}

func ToMorphGraphs(graphs []*Morph.BasicMorphGraph) []NLP.MorphDependencyGraph {
	morphs := make([]NLP.MorphDependencyGraph, len(graphs))
	for i, g := range graphs {
		morphs[i] = NLP.MorphDependencyGraph(g)
	}
	return morphs
}
