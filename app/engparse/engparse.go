package engparse

import (
	// "chukuparser/alg/featurevector"
	"chukuparser/alg/perceptron"
	"chukuparser/alg/search"
	"chukuparser/alg/transition"
	transitionmodel "chukuparser/alg/transition/model"
	"chukuparser/nlp/format/conll"
	"chukuparser/nlp/format/taggedsentence"
	"chukuparser/nlp/parser/dependency"
	. "chukuparser/nlp/parser/dependency/transition"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"
	"chukuparser/util/conf"

	"encoding/gob"
	"fmt"
	"log"
	"os"
	"runtime"
	// "strings"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func init() {
	gob.Register(&Serialization{})
}

var (
	allOut   bool = true
	parseOut bool = false

	// processing options
	Iterations, BeamSize int
	ConcurrentBeam       bool
	NumFeatures          int

	// global enumerations
	ERel, ETrans, EWord, EPOS, EWPOS, EMHost, EMSuffix *util.EnumSet

	// enumeration offsets of transitions
	SH, RE, PR, LA, RA transition.Transition

	// file names
	tConll       string
	input        string
	outConll     string
	modelFile    string
	featuresFile string
	labelsFile   string

	// command required flags
	REQUIRED_FLAGS []string = []string{"it", "tc", "in", "oc", "f", "l"}
)

// An approximation of the number of different MD-X:Y:Z transitions
// Pre-allocating the enumeration saves frequent reallocation during training and parsing
const (
	APPROX_WORDS, APPROX_POS        = 100, 100
	WORDS_POS_FACTOR                = 5
	APPROX_MHOSTS, APPROX_MSUFFIXES = 128, 16
)

func SetupRelationEnum(labels []string) {
	if ERel != nil {
		return
	}
	ERel = util.NewEnumSet(len(labels) + 1)
	ERel.Add(nlp.DepRel(nlp.ROOT_LABEL))
	for _, label := range labels {
		ERel.Add(nlp.DepRel(label))
	}
	ERel.Frozen = true
}

func SetupTransEnum(relations []string) {
	ETrans = util.NewEnumSet((len(relations)+1)*2 + 2)
	_, _ = ETrans.Add("NO") // dummy no action transition for zpar equivalence
	iSH, _ := ETrans.Add("SH")
	iRE, _ := ETrans.Add("RE")
	_, _ = ETrans.Add("AL") // dummy action transition for zpar equivalence
	_, _ = ETrans.Add("AR") // dummy action transition for zpar equivalence
	iPR, _ := ETrans.Add("PR")
	SH = transition.Transition(iSH)
	RE = transition.Transition(iRE)
	PR = transition.Transition(iPR)
	LA = PR + 1
	ETrans.Add("LA-" + string(nlp.ROOT_LABEL))
	for _, transition := range relations {
		ETrans.Add("LA-" + string(transition))
	}
	RA = transition.Transition(ETrans.Len())
	ETrans.Add("RA-" + string(nlp.ROOT_LABEL))
	for _, transition := range relations {
		ETrans.Add("RA-" + string(transition))
	}
}

func SetupEnum(relations []string) {
	SetupRelationEnum(relations)
	SetupTransEnum(relations)
	EWord, EPOS, EWPOS = util.NewEnumSet(APPROX_WORDS), util.NewEnumSet(APPROX_POS), util.NewEnumSet(APPROX_WORDS*WORDS_POS_FACTOR)
	EMHost, EMSuffix = util.NewEnumSet(APPROX_MHOSTS), util.NewEnumSet(APPROX_MSUFFIXES)
	// adding empty string as an element in the morph enum sets so that '0' default values
	// map to empty morphs
	EMHost.Add("")
	EMSuffix.Add("")
}

func SetupExtractor(setup *transition.FeatureSetup) *transition.GenericExtractor {
	extractor := &transition.GenericExtractor{
		EFeatures:  util.NewEnumSet(setup.NumFeatures()),
		Concurrent: false,
		EWord:      EWord,
		EPOS:       EPOS,
		EWPOS:      EWPOS,
		ERel:       ERel,
		EMHost:     EMHost,
		EMSuffix:   EMSuffix,
		// Log:        true,
	}
	extractor.Init()
	extractor.LoadFeatureSetup(setup)
	// for _, feature := range features {
	// 	featurePair := strings.Split(feature, ",")
	// 	if err := extractor.LoadFeature(featurePair[0], featurePair[1]); err != nil {
	// 		log.Fatalln("Failed to load feature", err.Error())
	// 	}
	// }
	NumFeatures = setup.NumFeatures()
	return extractor
}

func TrainingSequences(trainingSet []nlp.LabeledDependencyGraph, transitionSystem transition.TransitionSystem, extractor perceptron.FeatureExtractor) []perceptron.DecodedInstance {
	// verify feature load

	// mconf := &SimpleConfiguration{
	// 	EWord:  EWord,
	// 	EPOS:   EPOS,
	// 	EWPOS:  EWPOS,
	// 	ERel:   ERel,
	// 	ETrans: ETrans,
	// }
	// deterministic := &Deterministic{
	// 	TransFunc:          transitionSystem,
	// 	FeatExtractor:      extractor,
	// 	ReturnModelValue:   false,
	// 	ReturnSequence:     true,
	// 	ShowConsiderations: false,
	// 	Base:               mconf,
	// 	NoRecover:          false,
	// }

	// model := transitionmodel.NewAvgMatrixSparse(NumFeatures, nil)

	// tempModel := dependency.TransitionParameterModel(&PerceptronModel{model})

	instances := make([]perceptron.DecodedInstance, 0, len(trainingSet))
	// var failedTraining int
	for _, graph := range trainingSet {
		// if i%100 == 0 {
		// 	if allOut {
		// 		log.Println("At line", i)
		// 	}
		// 	runtime.GC()
		// }
		sent := graph.TaggedSentence()

		decoded := &perceptron.Decoded{sent, graph}
		instances = append(instances, decoded)

		// 	_, goldParams := deterministic.ParseOracle(graph, nil, tempModel)
		// 	if goldParams != nil {
		// 		seq := goldParams.(*ParseResultParameters).Sequence

		// 		goldSequence := make(ScoredConfigurations, len(seq))
		// 		var (
		// 			lastFeatures *transition.FeaturesList
		// 			curFeats     []featurevector.Feature
		// 		)
		// 		for i := len(seq) - 1; i >= 0; i-- {
		// 			val := seq[i]
		// 			curFeats = extractor.Features(val)
		// 			lastFeatures = &transition.FeaturesList{curFeats, val.GetLastTransition(), lastFeatures}
		// 			goldSequence[len(seq)-i-1] = &ScoredConfiguration{val.(DependencyConfiguration), val.GetLastTransition(), 0.0, lastFeatures, 0, 0, true}
		// 		}

		// 		// log.Println("Gold seq:\n", seq)
		// 		decoded := &perceptron.Decoded{sent, goldSequence}
		// 		instances = append(instances, decoded)
		// 	} else {
		// 		failedTraining++
		// 	}
	}
	return instances
}

func EstimatedBeamTransitions() int {
	return ERel.Len()*2 + 2
}

func Train(trainingSet []perceptron.DecodedInstance, Iterations, BeamSize int, filename string, model perceptron.Model, transitionSystem transition.TransitionSystem, extractor perceptron.FeatureExtractor) *perceptron.LinearPerceptron {
	conf := &SimpleConfiguration{
		EWord:    EWord,
		EPOS:     EPOS,
		EWPOS:    EWPOS,
		EMHost:   EMHost,
		EMSuffix: EMSuffix,
		ERel:     ERel,
		ETrans:   ETrans,
	}

	deterministic := &search.Deterministic{
		TransFunc:          transitionSystem,
		FeatExtractor:      extractor,
		ReturnModelValue:   false,
		ReturnSequence:     true,
		ShowConsiderations: false,
		Base:               conf,
		// NoRecover:          true,
	}

	beam := &search.Beam{
		TransFunc:            transitionSystem,
		FeatExtractor:        extractor,
		Base:                 conf,
		Size:                 BeamSize,
		ConcurrentExec:       ConcurrentBeam,
		EstimatedTransitions: EstimatedBeamTransitions(),
	}

	decoder := perceptron.EarlyUpdateInstanceDecoder(beam)
	golddec := perceptron.InstanceDecoder(deterministic)
	updater := new(transitionmodel.AveragedModelStrategy)

	perceptron := &perceptron.LinearPerceptron{
		Decoder:     decoder,
		GoldDecoder: golddec,
		Updater:     updater,
		Tempfile:    filename,
		TempLines:   1000}

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

func Parse(sents []nlp.EnumTaggedSentence, BeamSize int, model dependency.TransitionParameterModel, transitionSystem transition.TransitionSystem, extractor perceptron.FeatureExtractor) []nlp.LabeledDependencyGraph {
	conf := &SimpleConfiguration{
		EWord:    EWord,
		EPOS:     EPOS,
		EWPOS:    EWPOS,
		EMHost:   EMHost,
		EMSuffix: EMSuffix,
		ERel:     ERel,
		ETrans:   ETrans,
	}
	// runtime.GOMAXPROCS(1)
	beam := search.Beam{
		TransFunc:      transitionSystem,
		FeatExtractor:  extractor,
		Base:           conf,
		Size:           BeamSize,
		Model:          model,
		ConcurrentExec: ConcurrentBeam,
		// ConcurrentExec:  false,
		ShortTempAgenda:      true,
		EstimatedTransitions: EstimatedBeamTransitions(),
	}

	// Search.AllOut = true
	parsedGraphs := make([]nlp.LabeledDependencyGraph, len(sents))
	for i, sent := range sents {
		if i%5 == 0 {
			runtime.GC()
		}
		log.Println("Parsing sent", i) //, "len", len(sent.Tokens()))
		// }
		graph, _ := beam.Parse(sent)
		labeled := graph.(nlp.LabeledDependencyGraph)
		parsedGraphs[i] = labeled
	}
	if allOut {
		// log.Println("PARSE Total Time:", beam.DurTotal)
	}
	return parsedGraphs
}

func VerifyExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		log.Println("Error accessing file", filename)
		log.Println(err)
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

func ConfigOut(outModelFile string, b search.Interface, t transition.TransitionSystem) {
	log.Println("Configuration")
	log.Printf("Beam:             \t%s", b.Name())
	log.Printf("Transition System:\t%s", t.Name())
	log.Printf("Iterations:\t\t%d", Iterations)
	log.Printf("Beam Size:\t\t%d", BeamSize)
	log.Printf("Beam Concurrent:\t%v", ConcurrentBeam)
	log.Printf("Model file:\t\t%s", outModelFile)

	log.Println()
	log.Printf("Features File:\t%s", featuresFile)
	if !VerifyExists(featuresFile) {
		os.Exit(1)
	}
	log.Printf("Labels File:\t\t%s", labelsFile)
	if !VerifyExists(labelsFile) {
		os.Exit(1)
	}
	log.Println()
	log.Println("Data")
	log.Printf("Train file (conll):\t\t\t%s", tConll)
	if !VerifyExists(tConll) {
		os.Exit(1)
	}
	log.Printf("Test file  (tagged sentences):\t%s", input)
	if !VerifyExists(input) {
		os.Exit(1)
	}
	log.Printf("Out (conll) file:\t\t\t%s", outConll)
}

func EnglishTrainAndParse(cmd *commander.Command, args []string) {
	// arcSystem := &ArcEager{
	// 	ArcStandard: ArcStandard{
	// 		SHIFT:       SH,
	// 		LEFT:        LA,
	// 		RIGHT:       RA,
	// 		Relations:   ERel,
	// 		Transitions: ETrans,
	// 	},
	// 	REDUCE:  RE,
	// 	POPROOT: PR}
	arcSystem := &ArcStandard{
		SHIFT:       SH,
		LEFT:        LA,
		RIGHT:       RA,
		Relations:   ERel,
		Transitions: ETrans,
	}

	arcSystem.AddDefaultOracle()

	transitionSystem := transition.TransitionSystem(arcSystem)

	VerifyFlags(cmd)
	// RegisterTypes()
	var (
		outModelFile string                           = fmt.Sprintf("%s.b%d.i%d", modelFile, BeamSize, Iterations)
		model        *transitionmodel.AvgMatrixSparse = &transitionmodel.AvgMatrixSparse{}
	)
	if allOut && !parseOut {
		ConfigOut(outModelFile, &search.Beam{}, transitionSystem)
	}
	modelExists := VerifyExists(outModelFile)
	// modelExists := false
	relations, err := conf.ReadFile(labelsFile)
	if err != nil {
		log.Println("Failed reading dependency labels configuration file:", labelsFile)
		log.Fatalln(err)
	}
	if allOut && !parseOut {
		log.Println()
		// start processing - setup enumerations
		log.Println("Setup enumerations")
	}
	SetupEnum(relations.Values)

	if allOut && !parseOut {
		log.Println()
		log.Println("Loading features")
	}

	// features, err := conf.ReadFile(featuresFile)
	if err != nil {
		log.Println("Failed reading feature configuration file:", featuresFile)
		log.Fatalln(err)
	}
	featureSetup, err := transition.LoadFeatureConfFile(featuresFile)
	if err != nil {
		log.Println("Failed reading feature configuration file:", featuresFile)
		log.Fatalln(err)
	}
	extractor := SetupExtractor(featureSetup)
	// extractor.Log = true

	formatters := make([]util.Format, len(extractor.FeatureTemplates))
	for i, formatter := range extractor.FeatureTemplates {
		formatters[i] = formatter
	}

	if !modelExists {
		if allOut {
			log.Println("Model file", outModelFile, "not found, training")
		}
		if allOut {
			log.Println()

			log.Println("Generating Gold Sequences For Training")
			log.Println("Reading training sentences from", tConll)
		}
		s, e := conll.ReadFile(tConll)
		if e != nil {
			log.Fatalln(e)
		}
		// const NUM_SENTS = 20

		// s = s[:NUM_SENTS]
		if allOut {
			log.Println("Read", len(s), "sentences from", tConll)
			log.Println("Converting from conll to internal format")
		}
		goldGraphs := conll.Conll2GraphCorpus(s, EWord, EPOS, EWPOS, ERel, EMHost, EMSuffix)

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
		model = transitionmodel.NewAvgMatrixSparse(featureSetup.NumFeatures(), formatters, true)
		// model.Log = true
		_ = Train(goldSequences, Iterations, BeamSize, modelFile, model, transitionSystem, extractor)
		if allOut {
			log.Println("Done Training")
			log.Println()
			log.Println("Writing model to", outModelFile)
		}
		serialization := &Serialization{
			model.Serialize(),
			EWord, EPOS, EWPOS, EMHost, EMSuffix,
		}
		WriteModel(outModelFile, serialization)
		if allOut {
			log.Println("Done writing model")
		}
	} else {
		if allOut && !parseOut {
			log.Println("Found model file", outModelFile, " ... loading model")
		}
		serialization := ReadModel(outModelFile)
		model.Deserialize(serialization.WeightModel)
		EWord, EPOS, EWPOS, EMHost, EMSuffix = serialization.EWord, serialization.EPOS, serialization.EWPOS, serialization.EMHost, serialization.EMSuffix
		if allOut && !parseOut {
			log.Println("Loaded model")
		}
		// model.Log = true
	}
	if allOut {
		log.Println()
	}
	sents, e2 := taggedsentence.ReadFile(input, EWord, EPOS, EWPOS)
	formatters = make([]util.Format, len(extractor.FeatureTemplates))
	for i, _ := range extractor.FeatureTemplates {
		extractor.FeatureTemplates[i].EWord, extractor.FeatureTemplates[i].EPOS, extractor.FeatureTemplates[i].EWPOS = EWord, EPOS, EWPOS
		formatters[i] = &(extractor.FeatureTemplates[i])
	}

	model.Formatters = formatters
	// sents = sents[:NUM_SENTS]
	if allOut {
		if !parseOut {
			log.Println("Read", len(sents), "from", input)
		}
		if e2 != nil {
			log.Fatalln(e2)
		}
		if parseOut {
			log.SetPrefix("")
			log.SetFlags(0)
			log.Print("Parsing started")
		} else {
			log.Print("Parsing")
		}
		parsedGraphs := Parse(sents, BeamSize, model, arcSystem, extractor)
		if !parseOut {
			log.Println("Converting to conll")
		}
		graphAsConll := conll.Graph2ConllCorpus(parsedGraphs)
		conll.WriteFile(outConll, graphAsConll)
		if !parseOut {
			log.Println("Wrote", len(parsedGraphs), "in conll format to", outConll)
		}
	} else {
		search.AllOut = true
		// runtime.GOMAXPROCS(1)
		model.Log = true
		search.AllOut = true
		log.SetPrefix("")
		log.SetFlags(0)
		log.Print("Parsing started")
		parsedGraphs := Parse(sents, BeamSize, model, arcSystem, extractor)
		graphAsConll := conll.Graph2ConllCorpus(parsedGraphs)
		conll.WriteFile(outConll, graphAsConll)
		// log.Println("Wrote", len(parsedGraphs), "in conll format to", outConll)
	}

	// // sents, e2 := taggedsentence.ReadFile(input, EWord, EPOS, EWPOS)
	// if allOut {
	// 	linesForParse, e2 := conll.ReadFile(input)
	// 	if e2 != nil {
	// 		log.Fatalln(e2)
	// 	}
	// 	// sents = sents[:NUM_SENTS]
	// 	log.Println("Read", len(linesForParse), "from", input)
	// 	log.Println("Converting from conll to internal format")
	// 	forStripping := conll.Conll2GraphCorpus(linesForParse, EWord, EPOS, EWPOS, ERel)
	// 	strippedGraphs := make([]nlp.EnumTaggedSentence, len(forStripping))
	// 	for i, val := range forStripping {
	// 		strippedGraphs[i] = val.TaggedSentence().(nlp.EnumTaggedSentence)
	// 	}
	// 	log.Print("Parsing")
	// 	parsedGraphs := Parse(strippedGraphs, BeamSize, dependency.TransitionParameterModel(&PerceptronModel{model}), arcSystem, extractor)
	// 	log.Println("Converting to conll")
	// 	graphAsConll := conll.Graph2ConllCorpus(parsedGraphs)
	// 	log.Println("Wrote", len(parsedGraphs), "in conll format to", outConll)
	// 	conll.WriteFile(outConll, graphAsConll)
	// }
}

func EnglishCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       EnglishTrainAndParse,
		UsageLine: "english <file options> [arguments]",
		Short:     "runs english dependency training and parsing",
		Long: `
runs english dependency training and parsing

	$ ./chukuparser english -f <features> -l <labels> -tc <conll> -in <input tagged> -oc <out conll> [options]

`,
		Flag: *flag.NewFlagSet("english", flag.ExitOnError),
	}
	cmd.Flag.BoolVar(&ConcurrentBeam, "bconc", false, "Concurrent Beam")
	cmd.Flag.IntVar(&Iterations, "it", 1, "Number of Perceptron Iterations")
	cmd.Flag.IntVar(&BeamSize, "b", 4, "Beam Size")
	cmd.Flag.StringVar(&modelFile, "m", "model", "Prefix for model file ({m}.b{b}.i{it}.model)")

	cmd.Flag.StringVar(&tConll, "tc", "", "Training Conll File")
	cmd.Flag.StringVar(&input, "in", "", "Test Tagged Sentences File")
	cmd.Flag.StringVar(&outConll, "oc", "", "Output Conll File")
	cmd.Flag.StringVar(&featuresFile, "f", "", "Features Configuration File")
	cmd.Flag.StringVar(&labelsFile, "l", "", "Dependency Labels Configuration File")
	return cmd
}

type Serialization struct {
	WeightModel                          *transitionmodel.AvgMatrixSparseSerialized
	EWord, EPOS, EWPOS, EMHost, EMSuffix *util.EnumSet
}

func WriteModel(file string, data *Serialization) {
	fObj, err := os.Create(file)
	if err != nil {
		log.Fatalln("Failed creating model file", file, err)
		return
	}
	writer := gob.NewEncoder(fObj)
	writer.Encode(data)
}

func ReadModel(file string) *Serialization {
	data := &Serialization{}
	fObj, err := os.Open(file)
	if err != nil {
		log.Fatalln("Failed reading model from", file, err)
		return nil
	}
	reader := gob.NewDecoder(fObj)
	reader.Decode(data)
	return data
}
