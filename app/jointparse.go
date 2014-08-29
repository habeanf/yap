package app

import (
	"chukuparser/alg/perceptron"
	"chukuparser/alg/search"
	"chukuparser/alg/transition"
	transitionmodel "chukuparser/alg/transition/model"
	"chukuparser/nlp/format/conll"
	"chukuparser/nlp/format/lattice"
	"chukuparser/nlp/format/mapping"
	"chukuparser/nlp/format/segmentation"
	. "chukuparser/nlp/parser/dependency/transition"
	"chukuparser/nlp/parser/dependency/transition/morph"
	"chukuparser/nlp/parser/disambig"
	"chukuparser/nlp/parser/joint"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"
	"chukuparser/util/conf"

	"fmt"
	"log"
	"os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var (
	JointStrategy, OracleStrategy string
	AlignBeam                     bool
)

func SetupEnum(relations []string) {
	SetupRelationEnum(relations)
	SetupMorphTransEnum(relations)
	EWord, EPOS, EWPOS = util.NewEnumSet(APPROX_WORDS), util.NewEnumSet(APPROX_POS), util.NewEnumSet(APPROX_WORDS*5)
	EMHost, EMSuffix = util.NewEnumSet(APPROX_MHOSTS), util.NewEnumSet(APPROX_MSUFFIXES)
	EMorphProp = util.NewEnumSet(130) // random guess of number of possible values
	ETokens = util.NewEnumSet(10000)  // random guess of number of possible values
	// adding empty string as an element in the morph enum sets so that '0' default values
	// map to empty morphs
	EMHost.Add("")
	EMSuffix.Add("")
}

func CombineJointCorpus(graphs, goldLats, ambLats []interface{}) ([]interface{}, int) {
	if len(graphs) != len(goldLats) || len(graphs) != len(ambLats) {
		panic(fmt.Sprintf("Got mismatched training slice inputs (graphs, gold lattices, ambiguous lattices):", len(graphs), len(goldLats), len(ambLats)))
	}
	morphGraphs := make([]interface{}, len(graphs))
	var (
		numLatticeNoGold int
		noGold           bool
	)
	prefix := log.Prefix()
	for i, goldGraph := range graphs {
		goldLat := goldLats[i].(nlp.LatticeSentence)
		ambLat := ambLats[i].(nlp.LatticeSentence)
		log.SetPrefix(fmt.Sprintf("%v graph# %v ", prefix, i))
		morphGraphs[i], noGold = morph.CombineToGoldMorph(goldGraph.(nlp.LabeledDependencyGraph), goldLat, ambLat)
		if noGold {
			numLatticeNoGold++
		}
	}
	log.SetPrefix(prefix)
	return morphGraphs, numLatticeNoGold
}

func CombineToGoldMorphs(goldLats, ambLats []interface{}) ([]interface{}, int) {
	if len(goldLats) != len(ambLats) {
		panic(fmt.Sprintf("Got mismatched training slice inputs (gold lattices, ambiguous lattices):", len(goldLats), len(ambLats)))
	}
	morphGraphs := make([]interface{}, len(goldLats))
	var (
		numLatticeNoGold int
		noGold           bool
	)
	prefix := log.Prefix()
	for i, goldLat := range goldLats {
		ambLat := ambLats[i].(nlp.LatticeSentence)
		log.SetPrefix(fmt.Sprintf("%v lattice# %v ", prefix, i))
		morphGraphs[i], noGold = CombineToGoldMorph(goldLat.(nlp.LatticeSentence), ambLat)
		if noGold {
			numLatticeNoGold++
		}
	}
	log.SetPrefix(prefix)
	return morphGraphs, numLatticeNoGold
}

func JointConfigOut(outModelFile string, b search.Interface, t transition.TransitionSystem) {
	log.Println("*** CONFIGURATION ***")
	log.Printf("Beam:             \t%s", b.Name())
	log.Printf("Transition System:\t%s", t.Name())
	log.Printf("Transition Oracle:\t%s", t.Oracle().Name())
	log.Printf("Iterations:\t\t%d", Iterations)
	log.Printf("Beam Size:\t\t%d", BeamSize)
	log.Printf("Beam Concurrent:\t%v", ConcurrentBeam)
	log.Printf("Parameter Func:\t%v", paramFuncName)
	// log.Printf("Model file:\t\t%s", outModelFile)

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
		return
	}
	log.Printf("Train file (disamb. lattice):\t%s", tLatDis)
	if !VerifyExists(tLatDis) {
		return
	}
	log.Printf("Train file (ambig.  lattice):\t%s", tLatAmb)
	if !VerifyExists(tLatAmb) {
		return
	}
	log.Printf("Test file  (ambig.  lattice):\t%s", input)
	if !VerifyExists(input) {
		return
	}
	if len(inputGold) > 0 {
		log.Printf("Test file  (disambig.  lattice):\t%s", inputGold)
		if !VerifyExists(inputGold) {
			return
		}
	}
	log.Printf("Out (disamb.) file:\t\t\t%s", outLat)
	log.Printf("Out (segmt.) file:\t\t\t%s", outSeg)
	log.Printf("Out (mapping.) file:\t\t\t%s", outMap)
	log.Printf("Out Train (segmt.) file:\t\t%s", tSeg)
}

func JointTrainAndParse(cmd *commander.Command, args []string) {
	// *** SETUP ***
	paramFunc, exists := disambig.MDParams[paramFuncName]
	if !exists {
		log.Fatalln("Param Func", paramFuncName, "does not exist")
	}

	mdTrans := &disambig.MDTrans{
		ParamFunc: paramFunc,
	}

	arcSystem := &ArcStandard{
		SHIFT:       SH,
		LEFT:        LA,
		RIGHT:       RA,
		Relations:   ERel,
		Transitions: ETrans,
	}

	arcSystem.AddDefaultOracle()

	jointTrans := &joint.JointTrans{
		MDTrans:       mdTrans,
		ArcSys:        arcSystem,
		JointStrategy: JointStrategy,
	}
	jointTrans.AddDefaultOracle()
	jointTrans.Oracle().(*joint.JointOracle).OracleStrategy = OracleStrategy
	transitionSystem := transition.TransitionSystem(jointTrans)

	REQUIRED_FLAGS := []string{"it", "tc", "td", "tl", "in", "oc", "om", "os", "ots", "f", "l", "jointstr", "oraclestr"}

	VerifyFlags(cmd, REQUIRED_FLAGS)
	// RegisterTypes()

	outModelFile := fmt.Sprintf("%s.b%d.i%d", modelFile, BeamSize, Iterations)

	JointConfigOut(outModelFile, &search.Beam{Align: AlignBeam}, transitionSystem)

	relations, err := conf.ReadFile(labelsFile)
	if err != nil {
		log.Println("Failed reading dependency labels configuration file:", labelsFile)
		log.Fatalln(err)
	}
	if allOut {
		log.Println()
		// start processing - setup enumerations
		log.Println("Setup enumerations")
	}
	SetupEnum(relations.Values)

	arcSystem = &ArcStandard{
		SHIFT:       SH,
		LEFT:        LA,
		RIGHT:       RA,
		Relations:   ERel,
		Transitions: ETrans,
	}
	arcSystem.AddDefaultOracle()
	jointTrans.ArcSys = arcSystem
	jointTrans.Transitions = ETrans
	mdTrans.Transitions = ETrans
	mdTrans.AddDefaultOracle()
	jointTrans.MDTransition = MD
	jointTrans.JointStrategy = JointStrategy
	jointTrans.AddDefaultOracle()
	jointTrans.Oracle().(*joint.JointOracle).OracleStrategy = OracleStrategy

	transitionSystem = transition.TransitionSystem(jointTrans)

	if allOut {
		log.Println()
		log.Println("Loading features")
	}
	featureSetup, err := transition.LoadFeatureConfFile(featuresFile)
	if err != nil {
		log.Println("Failed reading feature configuration file:", featuresFile)
		log.Fatalln(err)
	}
	extractor := SetupExtractor(featureSetup)

	log.Println("")
	log.Println("*** TRAINING ***")
	// *** TRAINING ***

	if allOut {
		log.Println("Generating Gold Sequences For Training")
		log.Println("Conll:\tReading training conll sentences from", tConll)
	}
	s, e := conll.ReadFile(tConll)
	if e != nil {
		log.Println(e)
		return
	}
	if allOut {
		log.Println("Conll:\tRead", len(s), "sentences")
		log.Println("Conll:\tConverting from conll to internal structure")
	}
	goldConll := conll.Conll2GraphCorpus(s, EWord, EPOS, EWPOS, ERel, EMHost, EMSuffix)

	if allOut {
		log.Println("Dis. Lat.:\tReading training disambiguated lattices from", tLatDis)
	}
	lDis, lDisE := lattice.ReadFile(tLatDis)
	if lDisE != nil {
		log.Println(lDisE)
		return
	}
	if allOut {
		log.Println("Dis. Lat.:\tRead", len(lDis), "disambiguated lattices")
		log.Println("Dis. Lat.:\tConverting lattice format to internal structure")
	}
	goldDisLat := lattice.Lattice2SentenceCorpus(lDis, EWord, EPOS, EWPOS, EMorphProp, EMHost, EMSuffix)

	if allOut {
		log.Println("Amb. Lat:\tReading ambiguous lattices from", tLatAmb)
	}
	lAmb, lAmbE := lattice.ReadFile(tLatAmb)
	if lAmbE != nil {
		log.Println(lAmbE)
		return
	}
	if allOut {
		log.Println("Amb. Lat:\tRead", len(lAmb), "ambiguous lattices")
		log.Println("Amb. Lat:\tConverting lattice format to internal structure")
	}
	goldAmbLat := lattice.Lattice2SentenceCorpus(lAmb, EWord, EPOS, EWPOS, EMorphProp, EMHost, EMSuffix)
	if allOut {
		log.Println("Combining train files into gold morph graphs with original lattices")
	}
	combined, missingGold := CombineJointCorpus(goldConll, goldDisLat, goldAmbLat)

	if allOut {
		log.Println("Combined", len(combined), "graphs, with", missingGold, "missing at least one gold path in lattice")

		log.Println()

	}

	if allOut {
		log.Println()

		log.Println("Parsing with gold to get training sequences")
	}

	// const NUM_SENTS = 20
	// combined = combined[:NUM_SENTS]
	goldSequences := TrainingSequences(combined, GetMorphGraphAsLattices, GetMorphGraph)
	if allOut {
		log.Println("Generated", len(goldSequences), "training sequences")
		log.Println()
		// util.LogMemory()
		log.Println("Training", Iterations, "iteration(s)")
	}
	formatters := make([]util.Format, len(extractor.FeatureTemplates))
	for i, formatter := range extractor.FeatureTemplates {
		formatters[i] = formatter
	}
	model := transitionmodel.NewAvgMatrixSparse(NumFeatures, formatters, true)
	model.Extractor = extractor
	model.Classifier = func(t transition.Transition) string {
		if t < MD {
			return "Arc"
		} else {
			return "MD"
		}
	}

	conf := &joint.JointConfig{
		SimpleConfiguration: SimpleConfiguration{
			EWord:    EWord,
			EPOS:     EPOS,
			EWPOS:    EWPOS,
			EMHost:   EMHost,
			EMSuffix: EMSuffix,
			ERel:     ERel,
			ETrans:   ETrans,
		},
		MDConfig: disambig.MDConfig{
			ETokens: ETokens,
		},
	}

	beam := &search.Beam{
		TransFunc:            transitionSystem,
		FeatExtractor:        extractor,
		Base:                 conf,
		Size:                 BeamSize,
		ConcurrentExec:       ConcurrentBeam,
		Transitions:          ETrans,
		EstimatedTransitions: 1000,
		Align:                AlignBeam,
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

	_ = Train(goldSequences, Iterations, modelFile, model, perceptron.EarlyUpdateInstanceDecoder(beam), perceptron.InstanceDecoder(deterministic))
	search.AllOut = false
	if allOut {
		log.Println("Done Training")
		// util.LogMemory()
		log.Println()
		// log.Println("Writing final model to", outModelFile)
		// WriteModel(model, outModelFile)
	}

	// *** PARSING ***
	log.Println()
	log.Println("*** PARSING ***")
	log.Print("Parsing test")

	log.Println("Reading ambiguous lattices from", input)

	lAmb, lAmbE = lattice.ReadFile(input)
	if lAmbE != nil {
		log.Println(lAmbE)
		return
	}
	// lAmb = lAmb[:NUM_SENTS]
	if allOut {
		log.Println("Read", len(lAmb), "ambiguous lattices from", input)
		log.Println("Converting lattice format to internal structure")
	}
	predAmbLat := lattice.Lattice2SentenceCorpus(lAmb, EWord, EPOS, EWPOS, EMorphProp, EMHost, EMSuffix)

	if len(inputGold) > 0 {
		log.Println("Reading test disambiguated lattice (for test ambiguous infusion)")
		lDis, lDisE = lattice.ReadFile(inputGold)
		if lDisE != nil {
			log.Println(lDisE)
			return
		}
		if allOut {
			log.Println("Test Gold Dis. Lat.:\tRead", len(lDis), "disambiguated lattices")
			log.Println("Test Gold Dis. Lat.:\tConverting lattice format to internal structure")
		}

		predDisLat := lattice.Lattice2SentenceCorpus(lDis, EWord, EPOS, EWPOS, EMorphProp, EMHost, EMSuffix)

		if allOut {
			log.Println("Infusing test's gold disambiguation into ambiguous lattice")
		}

		_, missingGold = CombineToGoldMorphs(predDisLat, predAmbLat)

		if allOut {
			log.Println("Combined", len(combined), "graphs, with", missingGold, "missing at least one gold path in lattice")

			log.Println()
		}
	}
	beam.Model = model
	beam.ShortTempAgenda = true
	parsedGraphs := Parse(predAmbLat, beam)

	if allOut {
		log.Println("Converting", len(parsedGraphs), "to conll")
	}
	graphAsConll := conll.MorphGraph2ConllCorpus(parsedGraphs)
	if allOut {
		log.Println("Writing to output file")
	}
	conll.WriteFile(outLat, graphAsConll)
	if allOut {
		log.Println("Wrote", len(graphAsConll), "in conll format to", outLat)

		log.Println("Writing to segmentation file")
	}
	segmentation.WriteFile(outSeg, parsedGraphs)
	if allOut {
		log.Println("Wrote", len(parsedGraphs), "in segmentation format to", outSeg)

		log.Println("Writing to mapping file")
	}
	mapping.WriteFile(outMap, GetInstances(parsedGraphs, GetJointMDConfig))
	if allOut {
		log.Println("Wrote", len(parsedGraphs), "in mapping format to", outMap)

		log.Println("Writing to gold segmentation file")
	}
	segmentation.WriteFile(tSeg, combined)
	if allOut {
		log.Println("Wrote", len(combined), "in segmentation format to", tSeg)
	}
}

func JointCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       JointTrainAndParse,
		UsageLine: "joint <file options> [arguments]",
		Short:     "runs joint morpho-syntactic training and parsing",
		Long: `
runs morpho-syntactic training and parsing

	$ ./chukuparser joint -tc <conll> -td <train disamb. lat> -tl <train amb. lat> -in <input lat> -oc <out lat> -om <out map> -os <out seg> -ots <out train seg> -jointstr <joint strategy> -oraclestr <oracle strategy> [options]

`,
		Flag: *flag.NewFlagSet("joint", flag.ExitOnError),
	}
	cmd.Flag.BoolVar(&ConcurrentBeam, "bconc", false, "Concurrent Beam")
	cmd.Flag.IntVar(&Iterations, "it", 1, "Number of Perceptron Iterations")
	cmd.Flag.IntVar(&BeamSize, "b", 4, "Beam Size")
	cmd.Flag.StringVar(&modelFile, "m", "model", "Prefix for model file ({m}.b{b}.i{it}.model)")

	cmd.Flag.StringVar(&tConll, "tc", "", "Training Conll File")
	cmd.Flag.StringVar(&tLatDis, "td", "", "Training Disambiguated Lattices File")
	cmd.Flag.StringVar(&tLatAmb, "tl", "", "Training Ambiguous Lattices File")
	cmd.Flag.StringVar(&input, "in", "", "Test Ambiguous Lattices File")
	cmd.Flag.StringVar(&inputGold, "ing", "", "Optional - Gold Test Lattices File (for infusion into test ambiguous)")
	cmd.Flag.StringVar(&outLat, "oc", "", "Output Conll File")
	cmd.Flag.StringVar(&outSeg, "os", "", "Output Segmentation File")
	cmd.Flag.StringVar(&outMap, "om", "", "Output Mapping File")
	cmd.Flag.StringVar(&tSeg, "ots", "", "Output Training Segmentation File")
	cmd.Flag.StringVar(&featuresFile, "f", "", "Features Configuration File")
	cmd.Flag.StringVar(&labelsFile, "l", "", "Dependency Labels Configuration File")
	cmd.Flag.StringVar(&paramFuncName, "p", "Funcs_Main_POS_Both_Prop", "Param Func types: ["+disambig.AllParamFuncNames+"]")
	cmd.Flag.StringVar(&JointStrategy, "jointstr", "MDFirst", "Joint Strategy: ["+joint.JointStrategies+"]")
	cmd.Flag.StringVar(&OracleStrategy, "oraclestr", "MDFirst", "Oracle Strategy: ["+joint.OracleStrategies+"]")
	cmd.Flag.BoolVar(&AlignBeam, "align", false, "Use Beam Alignment")
	return cmd
}
