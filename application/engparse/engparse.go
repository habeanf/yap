package engparse

import (
	"chukuparser/algorithm/featurevector"
	"chukuparser/algorithm/perceptron"
	"chukuparser/algorithm/transition"
	TransitionModel "chukuparser/algorithm/transition/model"
	"chukuparser/nlp/format/conll"
	"chukuparser/nlp/format/taggedsentence"
	"chukuparser/nlp/parser/dependency"
	. "chukuparser/nlp/parser/dependency/transition"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"

	"encoding/gob"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var (
	allOut bool = true

	RICH_FEATURES [][2]string
	LABELS        []nlp.DepRel

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
	_, _ = ETrans.Add("NO") // dummy no action transition for zpar equivalence
	iSH, _ := ETrans.Add("SH")
	iRE, _ := ETrans.Add("RE")
	_, _ = ETrans.Add("AL") // dummy action transition for zpar equivalence
	_, _ = ETrans.Add("AR") // dummy action transition for zpar equivalence
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

func TrainingSequences(trainingSet []nlp.LabeledDependencyGraph, transitionSystem Transition.TransitionSystem, extractor Perceptron.FeatureExtractor) []Perceptron.DecodedInstance {
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
	model := TransitionModel.NewAvgMatrixSparse(len(RICH_FEATURES), nil)

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

			goldSequence := make(ScoredConfigurations, len(seq))
			var (
				lastFeatures *Transition.FeaturesList
				curFeats     []FeatureVector.Feature
			)
			for i := len(seq) - 1; i >= 0; i-- {
				val := seq[i]
				curFeats = extractor.Features(val)
				lastFeatures = &Transition.FeaturesList{curFeats, val.GetLastTransition(), lastFeatures}
				goldSequence[len(seq)-i-1] = &ScoredConfiguration{val.(DependencyConfiguration), val.GetLastTransition(), 0.0, lastFeatures, 0, 0, true}
			}

			// log.Println("Gold seq:\n", seq)
			decoded := &Perceptron.Decoded{sent, goldSequence}
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
	// beam.Log = true
	perceptron.Train(trainingSet)
	if allOut {
		log.Println("TRAIN Total Time:", beam.DurTotal)
	}
	return perceptron
}

func Parse(sents []nlp.EnumTaggedSentence, BeamSize int, model Dependency.TransitionParameterModel, transitionSystem Transition.TransitionSystem, extractor Perceptron.FeatureExtractor) []nlp.LabeledDependencyGraph {
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

	// Search.AllOut = true
	parsedGraphs := make([]nlp.LabeledDependencyGraph, len(sents))
	for i, sent := range sents {
		// if i%100 == 0 {
		// runtime.GC()
		log.Println("Parsing sent", i) //, "len", len(sent.Tokens()))
		// }
		graph, _ := beam.Parse(sent, nil, model)
		labeled := graph.(nlp.LabeledDependencyGraph)
		parsedGraphs[i] = labeled
	}
	if allOut {
		log.Println("PARSE Total Time:", beam.DurTotal)
	}
	return parsedGraphs
}

func RegisterTypes() {
	gob.Register(Transition.ConfigurationSequence{})
	gob.Register(&BasicDepGraph{})
	gob.Register(&nlp.TaggedToken{})
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
		//		log.Println("Error accessing file", filename)
		//		log.Println(err.Error())
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
	//	log.Println("Configuration")
	log.Printf("Beam:             \tStatic Length")
	log.Printf("Transition System:\tArcEager")
	log.Printf("Features:         \tRich")
	log.Printf("Iterations:\t\t%d", Iterations)
	log.Printf("Beam Size:\t\t%d", BeamSize)
	log.Printf("Beam Concurrent:\t%v", ConcurrentBeam)
	log.Printf("Model file:\t\t%s", outModelFile)
	//	log.Println()
	//	log.Println("Data")

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
	if allOut {
		outModelFile := fmt.Sprintf("%s.b%d.i%d", modelFile, BeamSize, Iterations)

		ConfigOut(outModelFile)
	}
	if allOut {
		log.Println()
		// start processing - setup enumerations
		log.Println("Setup enumerations")
	}
	SetupEnum()
	if allOut {
		log.Println()

		log.Println("Generating Gold Sequences For Training")
		log.Println("Reading training sentences from", tConll)
	}
	s, e := Conll.ReadFile(tConll)
	if e != nil {
		log.Println(e)
		return
	}
	// const NUM_SENTS = 20

	// s = s[:NUM_SENTS]
	if allOut {
		log.Println("Read", len(s), "sentences from", tConll)
		log.Println("Converting from conll to internal format")
	}
	goldGraphs := Conll.Conll2GraphCorpus(s, EWord, EPOS, EWPOS, ERel)

	if allOut {
		log.Println("Loading features")
	}
	extractor := &GenericExtractor{
		EFeatures:  Util.NewEnumSet(len(RICH_FEATURES)),
		Concurrent: false,
		EWord:      EWord,
		EPOS:       EPOS,
		EWPOS:      EWPOS,
		ERel:       ERel,
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

	if allOut {
		log.Println()

		log.Println("Parsing with gold to get training sequences")
	}
	// goldGraphs = goldGraphs[:NUM_SENTS]
	goldSequences := TrainingSequences(goldGraphs, transitionSystem, extractor)
	if allOut {
		log.Println("Generated", len(goldSequences), "training sequences")
		log.Println()
		log.Println("Training", Iterations, "iteration(s)")
	}
	formatters := make([]Util.Format, len(extractor.FeatureTemplates))
	for i, formatter := range extractor.FeatureTemplates {
		formatters[i] = formatter
	}
	model := TransitionModel.NewAvgMatrixSparse(len(RICH_FEATURES), formatters)
	_ = Train(goldSequences, Iterations, BeamSize, modelFile, model, transitionSystem, extractor)
	if allOut {
		log.Println("Done Training")
		log.Println()
	}
	sents, e2 := TaggedSentence.ReadFile(input, EWord, EPOS, EWPOS)
	// sents = sents[:NUM_SENTS]
	if allOut {
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
