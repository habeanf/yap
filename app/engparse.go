package app

import (
	// "yap/alg/featurevector"
	"yap/alg/perceptron"
	"yap/alg/search"
	"yap/alg/transition"
	transitionmodel "yap/alg/transition/model"
	"yap/nlp/format/conll"
	. "yap/nlp/parser/dependency/transition"
	"yap/util"
	"yap/util/conf"

	"fmt"
	"log"
	"os"
	// "strings"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func SetupEngEnum(relations []string) {
	SetupRelationEnum(relations)
	SetupTransEnum(relations)
	EWord, EPOS, EWPOS = util.NewEnumSet(APPROX_WORDS), util.NewEnumSet(APPROX_POS), util.NewEnumSet(APPROX_WORDS*WORDS_POS_FACTOR)
	EMHost, EMSuffix = util.NewEnumSet(APPROX_MHOSTS), util.NewEnumSet(APPROX_MSUFFIXES)
	EMorphProp = util.NewEnumSet(130) // random guess of number of possible values
	// adding empty string as an element in the morph enum sets so that '0' default values
	// map to empty morphs
	EMHost.Add("")
	EMSuffix.Add("")
}

func EstimatedBeamTransitions() int {
	return ERel.Len()*2 + 2
}

func EngConfigOut(outModelFile string, b search.Interface, t transition.TransitionSystem) {
	log.Println("Configuration")
	log.Printf("Beam:             \t%s", b.Name())
	log.Printf("Transition System:\t%s", t.Name())
	log.Printf("Iterations:\t\t%d", Iterations)
	log.Printf("Beam Size:\t\t%d", BeamSize)
	log.Printf("Beam Concurrent:\t%v", ConcurrentBeam)
	log.Printf("Model file:\t\t%s", outModelFile)
	log.Printf("Use Lemmas:\t\t%v", !conll.IGNORE_LEMMA)
	log.Printf("Word Type:\t\t%v", conll.WORD_TYPE)

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
	// instantiate the arc system for config output only
	// it will be reinstantiated later on with struct values

	var (
		arcSystem transition.TransitionSystem
	)
	switch arcSystemStr {
	case "standard":
		arcSystem = &ArcStandard{}
	case "eager":
		arcSystem = &ArcEager{}
	default:
		panic("Unknown arc system")
	}

	arcSystem.AddDefaultOracle()

	transitionSystem := transition.TransitionSystem(arcSystem)
	REQUIRED_FLAGS := []string{"it", "tc", "in", "oc", "f", "l"}

	VerifyFlags(cmd, REQUIRED_FLAGS)
	// RegisterTypes()
	var (
		outModelFile string                           = fmt.Sprintf("%s.b%d.i%d", modelFile, BeamSize, Iterations)
		model        *transitionmodel.AvgMatrixSparse = &transitionmodel.AvgMatrixSparse{}
	)
	if allOut && !parseOut {
		EngConfigOut(outModelFile, &search.Beam{}, transitionSystem)
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
	SetupEngEnum(relations.Values)

	// after calling SetupEngEnum, enums are instantiated and set according to the relations
	// therefore we re-instantiate the arc system with the right parameters
	switch arcSystemStr {
	case "standard":
		arcSystem = &ArcStandard{
			SHIFT:       SH.Value(),
			LEFT:        LA.Value(),
			RIGHT:       RA.Value(),
			Transitions: ETrans,
			Relations:   ERel,
		}
	case "eager":
		arcSystem = &ArcEager{
			ArcStandard: ArcStandard{
				SHIFT:       SH.Value(),
				LEFT:        LA.Value(),
				RIGHT:       RA.Value(),
				Relations:   ERel,
				Transitions: ETrans,
			},
			REDUCE:  RE.Value(),
			POPROOT: PR.Value(),
		}
	default:
		panic("Unknown arc system")
	}

	arcSystem.AddDefaultOracle()

	transitionSystem = transition.TransitionSystem(arcSystem)

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
	extractor := SetupExtractor(featureSetup, nil)
	// extractor.Log = true
	group, _ := extractor.TransTypeGroups[transition.ConstTransition(0).Type()]
	formatters := make([]util.Format, len(group.FeatureTemplates))
	for i, formatter := range group.FeatureTemplates {
		formatters[i] = formatter
	}

	devi, e2 := conll.ReadFile(input, 0)
	if e2 != nil {
		log.Fatalln(e2)
	}
	// const NUM_SENTS = 20

	// s = s[:NUM_SENTS]
	if allOut {
		log.Println("Read", len(devi), "sentences from", input)
		log.Println("Converting from conll to internal format")
	}
	asGraphs := conll.Conll2GraphCorpus(devi, EWord, EPOS, EWPOS, ERel, EMHost, EMSuffix)

	sents := make([]interface{}, len(asGraphs))
	for i, instance := range asGraphs {
		sents[i] = GetAsTaggedSentence(instance)
	}
	modelExists = false
	if !modelExists {
		if allOut {
			log.Println("Model file", outModelFile, "not found, training")
		}
		if allOut {
			log.Println()

			log.Println("Generating Gold Sequences For Training")
			log.Println("Reading training sentences from", tConll)
		}
		s, e := conll.ReadFile(tConll, 0)
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
		goldSequences := TrainingSequences(goldGraphs, GetAsTaggedSentence, GetAsLabeledDepGraph)
		if allOut {
			log.Println("Generated", len(goldSequences), "training sequences")
			log.Println()
			log.Println("Training", Iterations, "iteration(s)")
		}
		model = transitionmodel.NewAvgMatrixSparse(featureSetup.NumFeatures(), formatters, true)
		// model.Log = true

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
			NoRecover:          false,
		}

		beam := &search.Beam{
			TransFunc:            transitionSystem,
			FeatExtractor:        extractor,
			Base:                 conf,
			Size:                 BeamSize,
			ConcurrentExec:       ConcurrentBeam,
			EstimatedTransitions: EstimatedBeamTransitions(),
		}

		var evaluator perceptron.StopCondition

		if len(inputGold) > 0 {
			if allOut {
				log.Println("Setting convergence tester")
			}
			decodeTestBeam := &search.Beam{}
			*decodeTestBeam = *beam
			decodeTestBeam.Model = model
			decodeTestBeam.DecodeTest = true
			decodeTestBeam.ShortTempAgenda = true
			devigold, e3 := conll.ReadFile(inputGold, 0)
			if e3 != nil {
				log.Fatalln(e3)
			}
			if allOut {
				log.Println("Read", len(devigold), "sentences from", inputGold)
				log.Println("Converting from conll to internal format")
			}
			asGoldGraphs := conll.Conll2GraphCorpus(devigold, EWord, EPOS, EWPOS, ERel, EMHost, EMSuffix)

			goldSents := make([]interface{}, len(asGoldGraphs))
			for i, instance := range asGraphs {
				goldSents[i] = GetAsLabeledDepGraph(instance)
			}
			evaluator = MakeDepEvalStopCondition(sents, goldSents, decodeTestBeam, perceptron.InstanceDecoder(deterministic), BeamSize)
		}
		_ = Train(goldSequences, Iterations, modelFile, model, perceptron.EarlyUpdateInstanceDecoder(beam), perceptron.InstanceDecoder(deterministic), evaluator)
		if allOut {
			log.Println("Done Training")
			log.Println()
			log.Println("Writing model to", outModelFile)
		}
		serialization := &Serialization{
			model.Serialize(),
			EWord, EPOS, EWPOS, EMHost, EMSuffix, EMorphProp, ETrans,
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

	// lDisamb, lDisambE := lattice.ReadFile(input)
	// if lDisambE != nil {
	// 	log.Println(lDisambE)
	// 	return
	// }
	// // lDisamb = lDisamb[:NUM_SENTS]
	// if allOut {
	// 	log.Println("Read", len(lDisamb), "disambiguated lattices from", input)
	// 	log.Println("Converting lattice format to internal structure")
	// }
	// sents := lattice.Lattice2SentenceCorpus(lDisamb, EWord, EPOS, EWPOS, EMorphProp, EMHost, EMSuffix)

	group, _ = extractor.TransTypeGroups[transition.ConstTransition(0).Type()]
	formatters = make([]util.Format, len(group.FeatureTemplates))
	for i, _ := range group.FeatureTemplates {
		group.FeatureTemplates[i].EWord, group.FeatureTemplates[i].EPOS, group.FeatureTemplates[i].EWPOS = EWord, EPOS, EWPOS
		formatters[i] = &(group.FeatureTemplates[i])
	}

	model.Formatters = formatters
	// sents = sents[:NUM_SENTS]

	conf := &SimpleConfiguration{
		EWord:    EWord,
		EPOS:     EPOS,
		EWPOS:    EWPOS,
		EMHost:   EMHost,
		EMSuffix: EMSuffix,
		ERel:     ERel,
		ETrans:   ETrans,
	}

	beam := &search.Beam{
		TransFunc:            transitionSystem,
		FeatExtractor:        extractor,
		Base:                 conf,
		Model:                model,
		Size:                 BeamSize,
		ConcurrentExec:       ConcurrentBeam,
		ShortTempAgenda:      true,
		EstimatedTransitions: EstimatedBeamTransitions(),
	}
	if allOut {
		if !parseOut {
			log.Println("Read", len(sents), "from", input)
		}
		if parseOut {
			log.SetPrefix("")
			log.SetFlags(0)
			log.Print("Parsing started")
		} else {
			log.Print("Parsing")
		}

		parsedGraphs := Parse(sents, beam)
		if !parseOut {
			log.Println("Converting to conll")
		}
		graphAsConll := conll.Graph2ConllCorpus(parsedGraphs, EMHost, EMSuffix)
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
		parsedGraphs := Parse(sents, beam)
		graphAsConll := conll.Graph2ConllCorpus(parsedGraphs, EMHost, EMSuffix)
		conll.WriteFile(outConll, graphAsConll)
		log.Println("Wrote", len(parsedGraphs), "in conll format to", outConll)
	}
}

func EnglishCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       EnglishTrainAndParse,
		UsageLine: "english <file options> [arguments]",
		Short:     "runs english dependency training and parsing",
		Long: `
runs english dependency training and parsing

	$ ./yap english -f <features> -l <labels> -tc <conll> -in <input tagged> -oc <out conll> [-a eager|standard] [options]

`,
		Flag: *flag.NewFlagSet("english", flag.ExitOnError),
	}
	cmd.Flag.BoolVar(&ConcurrentBeam, "bconc", false, "Concurrent Beam")
	cmd.Flag.IntVar(&Iterations, "it", 1, "Number of Perceptron Iterations")
	cmd.Flag.IntVar(&BeamSize, "b", 4, "Beam Size")
	cmd.Flag.StringVar(&modelFile, "m", "model", "Prefix for model file ({m}.b{b}.i{it}.model)")
	cmd.Flag.StringVar(&arcSystemStr, "a", "eager", "Optional - Arc System [standard, eager]")

	cmd.Flag.StringVar(&tConll, "tc", "", "Training Conll File")
	cmd.Flag.StringVar(&input, "in", "", "Test Tagged Sentences File")
	cmd.Flag.StringVar(&inputGold, "ing", "", "Optional - Gold Parsed Sentences (for convergence)")
	cmd.Flag.StringVar(&outConll, "oc", "", "Output Conll File")
	cmd.Flag.StringVar(&featuresFile, "f", "", "Features Configuration File")
	cmd.Flag.StringVar(&labelsFile, "l", "", "Dependency Labels Configuration File")
	cmd.Flag.BoolVar(&conll.IGNORE_LEMMA, "nolemma", false, "Ignore lemmas")
	cmd.Flag.StringVar(&conll.WORD_TYPE, "wordtype", "lemma+f", "Word type [form, lemma, lemma+f (=lemma if present else form)]")
	return cmd
}
