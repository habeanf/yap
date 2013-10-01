package engparse

import (
	"chukuparser/Algorithm/Perceptron"
	"chukuparser/Algorithm/Transition"
	TransitionModel "chukuparser/Algorithm/Transition/Model"
	"chukuparser/NLP/Format/Conll"
	"chukuparser/NLP/Format/TaggedSentence"
	"chukuparser/NLP/Parser/Dependency"
	. "chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
	"chukuparser/Util"

	"encoding/gob"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var (
	RICH_FEATURES [][2]string = [][2]string{
		{"S0|w", "S0|w"},
		{"S0|p", "S0|w"},
		{"S0|w|p", "S0|w"},

		{"N0|w", "N0|w"},
		{"N0|p", "N0|w"},
		{"N0|w|p", "N0|w"},

		{"N1|w", "N1|w"},
		{"N1|p", "N1|w"},
		{"N1|w|p", "N1|w"},

		{"N2|w", "N2|w"},
		{"N2|p", "N2|w"},
		{"N2|w|p", "N2|w"},

		{"S0h|w", "S0h|w"},
		{"S0h|p", "S0h|w"},
		{"S0|l", "S0h|w"},

		{"S0h2|w", "S0h2|w"},
		{"S0h2|p", "S0h2|w"},
		{"S0h|l", "S0h2|w"},

		{"S0l|w", "S0l|w"},
		{"S0l|p", "S0l|w"},
		{"S0l|l", "S0l|w"},

		{"S0r|w", "S0r|w"},
		{"S0r|p", "S0r|w"},
		{"S0r|l", "S0r|w"},

		{"S0l2|w", "S0l2|w"},
		{"S0l2|p", "S0l2|w"},
		{"S0l2|l", "S0l2|w"},

		{"S0r2|w", "S0r2|w"},
		{"S0r2|p", "S0r2|w"},
		{"S0r2|l", "S0r2|w"},

		{"N0l|w", "N0l|w"},
		{"N0l|p", "N0l|w"},
		{"N0l|l", "N0l|w"},

		{"N0l2|w", "N0l2|w"},
		{"N0l2|p", "N0l2|w"},
		{"N0l2|l", "N0l2|w"},

		{"S0|w|p+N0|w|p", "S0|w"},
		{"S0|w|p+N0|w", "S0|w"},
		{"S0|w+N0|w|p", "S0|w"},
		{"S0|w|p+N0|p", "S0|w"},
		{"S0|p+N0|w|p", "S0|w"},
		{"S0|w+N0|w", "S0|w"},
		{"S0|p+N0|p", "S0|w"},

		{"N0|p+N1|p", "S0|w,N0|w"},
		{"N0|p+N1|p+N2|p", "S0|w,N0|w"},
		{"S0|p+N0|p+N1|p", "S0|w,N0|w"},
		{"S0|p+N0|p+N0l|p", "S0|w,N0|w"},
		{"N0|p+N0l|p+N0l2|p", "S0|w,N0|w"},

		{"S0h|p+S0|p+N0|p", "S0|w"},
		{"S0h2|p+S0h|p+S0|p", "S0|w"},
		{"S0|p+S0l|p+N0|p", "S0|w"},
		{"S0|p+S0l|p+S0l2|p", "S0|w"},
		{"S0|p+S0r|p+N0|p", "S0|w"},
		{"S0|p+S0r|p+S0r2|p", "S0|w"},

		{"S0|w|d", "S0|w,N0|w"},
		{"S0|p|d", "S0|w,N0|w"},
		{"N0|w|d", "S0|w,N0|w"},
		{"N0|p|d", "S0|w,N0|w"},
		{"S0|w+N0|w|d", "S0|w,N0|w"},
		{"S0|p+N0|p|d", "S0|w,N0|w"},

		{"S0|w|vr", "S0|w"},
		{"S0|p|vr", "S0|w"},
		{"S0|w|vl", "S0|w"},
		{"S0|p|vl", "S0|w"},
		{"N0|w|vl", "N0|w"},
		{"N0|p|vl", "N0|w"},

		{"S0|w|sr", "S0|w"},
		{"S0|p|sr", "S0|w"},
		{"S0|w|sl", "S0|w"},
		{"S0|p|sl", "S0|w"},
		{"N0|w|sl", "N0|w"},
		{"N0|p|sl", "N0|w"}}

	LABELS []NLP.DepRel = []NLP.DepRel{
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
		NLP.ROOT_LABEL,
	}

	Iterations     int
	BeamSize       int
	ConcurrentBeam bool
	tConll         string
	input          string
	outConll       string
	modelFile      string
	REQUIRED_FLAGS []string = []string{"it", "tc", "in", "oc"}

	// Global enumerations
	ERel, ETrans, EWord, EPOS, EWPOS *Util.EnumSet

	// Enumeration offsets of transitions
	SH, RE, PR, LA, RA Transition.Transition
)

func SetupRelationEnum() {
	if ERel != nil {
		return
	}
	ERel = Util.NewEnumSet(len(LABELS))
	for _, label := range LABELS {
		ERel.Add(label)
	}
	ERel.Frozen = true
}

// An approximation of the number of different MD-X:Y:Z transitions
// Pre-allocating the enumeration saves frequent reallocation during training and parsing
const (
	APPROX_WORDS, APPROX_POS = 100, 100
)

func SetupTransEnum() {
	ETrans = Util.NewEnumSet(len(LABELS)*2 + 2)
	iSH, _ := ETrans.Add("SH")
	iRE, _ := ETrans.Add("RE")
	iPR, _ := ETrans.Add("PR")
	SH = Transition.Transition(iSH)
	RE = Transition.Transition(iRE)
	PR = Transition.Transition(iPR)
	LA = PR + 1
	for _, transition := range LABELS {
		ETrans.Add("LA-" + string(transition))
	}
	RA = Transition.Transition(ETrans.Len())
	for _, transition := range LABELS {
		ETrans.Add("RA-" + string(transition))
	}
}

func SetupEnum() {
	SetupRelationEnum()
	SetupTransEnum()
	EWord, EPOS, EWPOS = Util.NewEnumSet(APPROX_WORDS), Util.NewEnumSet(APPROX_POS), Util.NewEnumSet(APPROX_WORDS*5)
}

func TrainingSequences(trainingSet []NLP.LabeledDependencyGraph, transitionSystem Transition.TransitionSystem, extractor Perceptron.FeatureExtractor) []Perceptron.DecodedInstance {
	// verify feature load

	mconf := &SimpleConfiguration{
		EWord:  EWord,
		EPOS:   EPOS,
		EWPOS:  EWPOS,
		ERel:   ERel,
		ETrans: ETrans,
	}
	deterministic := &Deterministic{
		TransFunc:          transitionSystem,
		FeatExtractor:      extractor,
		ReturnModelValue:   false,
		ReturnSequence:     true,
		ShowConsiderations: false,
		Base:               mconf,
		NoRecover:          true,
	}

	decoder := Perceptron.EarlyUpdateInstanceDecoder(deterministic)
	updater := new(TransitionModel.AveragedModelStrategy)

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	model := TransitionModel.NewAvgMatrixSparse(ETrans.Len(), len(RICH_FEATURES))

	tempModel := Dependency.TransitionParameterModel(&PerceptronModel{model})
	perceptron.Init(model)

	instances := make([]Perceptron.DecodedInstance, 0, len(trainingSet))
	var failedTraining int
	for i, graph := range trainingSet {
		if i%300 == 0 {
			log.Println("At line", i)
			runtime.GC()
		}
		sent := graph.TaggedSentence()

		_, goldParams := deterministic.ParseOracle(graph, nil, tempModel)
		if goldParams != nil {
			seq := goldParams.(*ParseResultParameters).Sequence
			// log.Println("Gold seq:\n", seq)
			decoded := &Perceptron.Decoded{sent, seq[0]}
			instances = append(instances, decoded)
		} else {
			failedTraining++
		}
	}
	return instances
}

func Train(trainingSet []Perceptron.DecodedInstance, Iterations, BeamSize int, filename string, model Perceptron.Model, transitionSystem Transition.TransitionSystem, extractor Perceptron.FeatureExtractor) *Perceptron.LinearPerceptron {
	conf := &SimpleConfiguration{
		EWord:  EWord,
		EPOS:   EPOS,
		EWPOS:  EWPOS,
		ERel:   ERel,
		ETrans: ETrans,
	}

	beam := &Beam{
		TransFunc:      transitionSystem,
		FeatExtractor:  extractor,
		Base:           conf,
		NumRelations:   len(LABELS),
		Size:           BeamSize,
		ConcurrentExec: ConcurrentBeam,
	}

	decoder := Perceptron.EarlyUpdateInstanceDecoder(beam)
	updater := new(TransitionModel.AveragedModelStrategy)

	perceptron := &Perceptron.LinearPerceptron{
		Decoder:   decoder,
		Updater:   updater,
		Tempfile:  filename,
		TempLines: 1000}

	perceptron.Iterations = Iterations
	perceptron.Init(model)
	// perceptron.TempLoad("model.b64.i1")
	perceptron.Log = true

	perceptron.Train(trainingSet)
	log.Println("TRAIN Total Time:", beam.DurTotal)
	log.Println("TRAIN Time Expanding (pct):\t", beam.DurExpanding.Seconds(), 100*beam.DurExpanding/beam.DurTotal)
	log.Println("TRAIN Time Inserting (pct):\t", beam.DurInserting.Seconds(), 100*beam.DurInserting/beam.DurTotal)
	log.Println("TRAIN Time Inserting-Feat (pct):\t", beam.DurInsertFeat.Seconds(), 100*beam.DurInsertFeat/beam.DurTotal)
	log.Println("TRAIN Time Inserting-Modl (pct):\t", beam.DurInsertModl.Seconds(), 100*beam.DurInsertModl/beam.DurTotal)
	log.Println("TRAIN Time Inserting-ModA (pct):\t", beam.DurInsertModA.Seconds(), 100*beam.DurInsertModA/beam.DurTotal)
	log.Println("TRAIN Time Inserting-Scrp (pct):\t", beam.DurInsertScrp.Seconds(), 100*beam.DurInsertScrp/beam.DurTotal)
	log.Println("TRAIN Time Inserting-Scrm (pct):\t", beam.DurInsertScrm.Seconds(), 100*beam.DurInsertScrm/beam.DurTotal)
	log.Println("TRAIN Time Inserting-Heap (pct):\t", beam.DurInsertHeap.Seconds(), 100*beam.DurInsertHeap/beam.DurTotal)
	log.Println("TRAIN Time Inserting-Agen (pct):\t", beam.DurInsertAgen.Seconds(), 100*beam.DurInsertAgen/beam.DurTotal)
	log.Println("TRAIN Time Inserting-Init (pct):\t", beam.DurInsertInit.Seconds(), 100*beam.DurInsertInit/beam.DurTotal)
	log.Println("TRAIN Time Top (pct):\t\t", beam.DurTop.Seconds(), 100*beam.DurTop/beam.DurTotal)
	log.Println("TRAIN Time TopB (pct):\t\t", beam.DurTopB.Seconds(), 100*beam.DurTopB/beam.DurTotal)
	log.Println("TRAIN Time Clearing (pct):\t\t", beam.DurClearing.Seconds(), 100*beam.DurClearing/beam.DurTotal)
	log.Println("TRAIN Total Time:", beam.DurTotal.Seconds())

	return perceptron
}

func Parse(sents []NLP.EnumTaggedSentence, BeamSize int, model Dependency.TransitionParameterModel, transitionSystem Transition.TransitionSystem, extractor Perceptron.FeatureExtractor) []NLP.LabeledDependencyGraph {
	conf := &SimpleConfiguration{
		EWord:  EWord,
		EPOS:   EPOS,
		EWPOS:  EWPOS,
		ERel:   ERel,
		ETrans: ETrans,
	}

	beam := Beam{
		TransFunc:       transitionSystem,
		FeatExtractor:   extractor,
		Base:            conf,
		Size:            BeamSize,
		NumRelations:    len(LABELS),
		Model:           model,
		ConcurrentExec:  ConcurrentBeam,
		ShortTempAgenda: true}

	parsedGraphs := make([]NLP.LabeledDependencyGraph, len(sents))
	for i, sent := range sents {
		// if i%100 == 0 {
		runtime.GC()
		log.Println("Parsing sent", i)
		// }
		graph, _ := beam.Parse(sent, nil, model)
		labeled := graph.(NLP.LabeledDependencyGraph)
		parsedGraphs[i] = labeled
	}
	log.Println("PARSE Time Expanding (pct):\t", beam.DurExpanding.Seconds(), 100*beam.DurExpanding/beam.DurTotal)
	log.Println("PARSE Time Inserting (pct):\t", beam.DurInserting.Seconds(), 100*beam.DurInserting/beam.DurTotal)
	log.Println("PARSE Time Inserting-Feat (pct):\t", beam.DurInsertFeat.Seconds(), 100*beam.DurInsertFeat/beam.DurTotal)
	log.Println("PARSE Time Inserting-Modl (pct):\t", beam.DurInsertModl.Seconds(), 100*beam.DurInsertModl/beam.DurTotal)
	log.Println("PARSE Time Inserting-ModA (pct):\t", beam.DurInsertModA.Seconds(), 100*beam.DurInsertModA/beam.DurTotal)
	log.Println("PARSE Time Inserting-Scrp (pct):\t", beam.DurInsertScrp.Seconds(), 100*beam.DurInsertScrp/beam.DurTotal)
	log.Println("PARSE Time Inserting-Scrm (pct):\t", beam.DurInsertScrm.Seconds(), 100*beam.DurInsertScrm/beam.DurTotal)
	log.Println("PARSE Time Inserting-Heap (pct):\t", beam.DurInsertHeap.Seconds(), 100*beam.DurInsertHeap/beam.DurTotal)
	log.Println("PARSE Time Inserting-Agen (pct):\t", beam.DurInsertAgen.Seconds(), 100*beam.DurInsertAgen/beam.DurTotal)
	log.Println("PARSE Time Inserting-Init (pct):\t", beam.DurInsertInit.Seconds(), 100*beam.DurInsertInit/beam.DurTotal)
	log.Println("PARSE Time Top (pct):\t\t", beam.DurTop.Seconds(), 100*beam.DurTop/beam.DurTotal)
	log.Println("PARSE Time TopB (pct):\t\t", beam.DurTopB.Seconds(), 100*beam.DurTopB/beam.DurTotal)
	log.Println("PARSE Time Clearing (pct):\t\t", beam.DurClearing.Seconds(), 100*beam.DurClearing/beam.DurTotal)
	log.Println("PARSE Total Time:", beam.DurTotal.Seconds())

	return parsedGraphs
}

func RegisterTypes() {
	gob.Register(Transition.ConfigurationSequence{})
	gob.Register(&BasicDepGraph{})
	gob.Register(&NLP.TaggedToken{})
	gob.Register(&BasicDepArc{})
	gob.Register(&Beam{})
	gob.Register(&SimpleConfiguration{})
	gob.Register(&ArcEager{})
	gob.Register(&GenericExtractor{})
	gob.Register(&PerceptronModel{})
	gob.Register(&Perceptron.AveragedStrategy{})
	gob.Register(&Perceptron.Decoded{})
	// gob.Register(TaggedSentence{})
	gob.Register(&StackArray{})
	gob.Register(&ArcSetSimple{})
	gob.Register([3]interface{}{})
	gob.Register(new(Transition.Transition))
}

func VerifyExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		log.Println("Error accessing file", filename)
		log.Println(err.Error())
		return false
	}
	return true
}

func VerifyFlags(cmd *commander.Command) {
	for _, flag := range REQUIRED_FLAGS {
		f := cmd.Flag.Lookup(flag)
		if f.Value.String() == "" {
			log.Printf("Required flag %s not set", f.Name)
			cmd.Usage()
			os.Exit(1)
		}
	}
}

func ConfigOut(outModelFile string) {
	log.Println("Configuration")
	log.Printf("Beam:             \tStatic Length")
	log.Printf("Transition System:\tArcEager")
	log.Printf("Features:         \tRich")
	log.Printf("Iterations:\t\t%d", Iterations)
	log.Printf("Beam Size:\t\t%d", BeamSize)
	log.Printf("Beam Concurrent:\t%v", ConcurrentBeam)
	log.Printf("Model file:\t\t%s", outModelFile)
	log.Println()
	log.Println("Data")

	log.Printf("Train file (conll):\t\t\t%s", tConll)
	if !VerifyExists(tConll) {
		return
	}
	log.Printf("Test file  (tagged sentences):\t%s", input)
	if !VerifyExists(input) {
		return
	}
	log.Printf("Out (conll) file:\t\t\t%s", outConll)
}

func EnglishTrainAndParse(cmd *commander.Command, args []string) {
	VerifyFlags(cmd)
	RegisterTypes()

	outModelFile := fmt.Sprintf("%s.b%d.i%d", modelFile, BeamSize, Iterations)

	ConfigOut(outModelFile)

	log.Println()
	// start processing - setup enumerations
	log.Println("Setup enumerations")
	SetupEnum()
	log.Println()

	log.Println("Generating Gold Sequences For Training")
	log.Println("Reading training sentences from", tConll)
	s, e := Conll.ReadFile(tConll)
	if e != nil {
		log.Println(e)
		return
	}
	// const NUM_SENTS = 20

	// s = s[:NUM_SENTS]
	log.Println("Read", len(s), "sentences from", tConll)
	log.Println("Converting from conll to internal format")
	goldGraphs := Conll.Conll2GraphCorpus(s, EWord, EPOS, EWPOS, ERel)

	log.Println("Loading features")
	extractor := &GenericExtractor{
		EFeatures:  Util.NewEnumSet(len(RICH_FEATURES)),
		Concurrent: false,
	}
	extractor.Init()
	for _, featurePair := range RICH_FEATURES {
		if err := extractor.LoadFeature(featurePair[0], featurePair[1]); err != nil {
			log.Panicln("Failed to load feature", err.Error())
		}
	}

	arcSystem := &ArcEager{
		ArcStandard: ArcStandard{
			SHIFT:       SH,
			LEFT:        LA,
			RIGHT:       RA,
			Relations:   ERel,
			Transitions: ETrans,
		},
		REDUCE:  RE,
		POPROOT: PR}
	arcSystem.AddDefaultOracle()

	transitionSystem := Transition.TransitionSystem(arcSystem)

	log.Println()

	log.Println("Parsing with gold to get training sequences")
	// goldGraphs = goldGraphs[:NUM_SENTS]
	goldSequences := TrainingSequences(goldGraphs, transitionSystem, extractor)
	log.Println("Generated", len(goldSequences), "training sequences")
	log.Println()
	log.Println("Training ( concurrent = ", ConcurrentBeam, ")", Iterations, "iteration(s)")
	model := TransitionModel.NewAvgMatrixSparse(ETrans.Len(), len(RICH_FEATURES))
	_ = Train(goldSequences, Iterations, BeamSize, modelFile, model, transitionSystem, extractor)
	log.Println("Done Training")
	log.Println()

	sents, e2 := TaggedSentence.ReadFile(input, EWord, EPOS, EWPOS)
	// sents = sents[:NUM_SENTS]
	log.Println("Read", len(sents), "from", input)
	if e2 != nil {
		log.Println(e2)
		return
	}

	log.Print("Parsing")
	parsedGraphs := Parse(sents, BeamSize, Dependency.TransitionParameterModel(&PerceptronModel{model}), arcSystem, extractor)
	log.Println("Converting to conll")
	graphAsConll := Conll.Graph2ConllCorpus(parsedGraphs)
	log.Println("Wrote", len(parsedGraphs), "in conll format to", outConll)
	Conll.WriteFile(outConll, graphAsConll)
}

func EnglishCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       EnglishTrainAndParse,
		UsageLine: "english <file options> [arguments]",
		Short:     "runs english dependency training and parsing",
		Long: `
runs english dependency training and parsing

	$ ./chukuparser english -tc <conll> -in <input tagged> -oc <out conll> [options]

`,
		Flag: *flag.NewFlagSet("english", flag.ExitOnError),
	}
	cmd.Flag.BoolVar(&ConcurrentBeam, "bconc", false, "Concurrent Beam")
	cmd.Flag.IntVar(&Iterations, "it", 1, "Number of Perceptron Iterations")
	cmd.Flag.IntVar(&BeamSize, "b", 4, "Beam Size")
	cmd.Flag.StringVar(&modelFile, "m", "model", "Prefix for model file ({m}.b{b}.i{i}.model)")

	cmd.Flag.StringVar(&tConll, "tc", "", "Training Conll File")
	cmd.Flag.StringVar(&input, "in", "", "Test Tagged Sentences File")
	cmd.Flag.StringVar(&outConll, "oc", "", "Output Conll File")
	return cmd
}
