package app

import (
	"yap/alg/perceptron"
	"yap/alg/search"
	"yap/alg/transition"
	transitionmodel "yap/alg/transition/model"
	"yap/nlp/format/lattice"
	// "yap/nlp/format/mapping"
	"yap/nlp/parser/disambig"

	nlp "yap/nlp/types"
	"yap/util"

	"fmt"
	"log"
	"os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var (
	paramFuncName string
	UseWB         bool
)

func SetupMDEnum() {
	EWord, EPOS, EWPOS = util.NewEnumSet(APPROX_WORDS), util.NewEnumSet(APPROX_POS), util.NewEnumSet(APPROX_WORDS*5)
	EMHost, EMSuffix = util.NewEnumSet(APPROX_MHOSTS), util.NewEnumSet(APPROX_MSUFFIXES)

	ETrans = util.NewEnumSet(10000)
	_, _ = ETrans.Add("IDLE")    // dummy no action transition for zpar equivalence
	iPOP, _ := ETrans.Add("POP") // dummy no action transition for zpar equivalence

	POP = transition.Transition(iPOP)

	EMorphProp = util.NewEnumSet(130) // random guess of number of possible values
	ETokens = util.NewEnumSet(10000)
}

func CombineToGoldMorph(goldLat, ambLat nlp.LatticeSentence) (m *disambig.MDConfig, addedMissingSpellout bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered error", r, "excluding from training corpus")
			m = nil
		}
	}()
	// generate graph

	// generate morph. disambiguation (= mapping) and nodes
	mappings := make([]*nlp.Mapping, len(goldLat))
	for i, lat := range goldLat {
		// log.Println("At lat", i)
		lat.GenSpellouts()
		lat.GenToken()
		if len(lat.Spellouts) == 0 {
			continue
		}
		mapping := &nlp.Mapping{
			lat.Token,
			lat.Spellouts[0],
		}
		// if the gold spellout doesn't exist in the lattice, add it
		_, exists := ambLat[i].Spellouts.Find(mapping.Spellout)
		if !exists {
			// log.Println(mapping.Spellout, "Spellout not found")
			ambLat[i].Spellouts = append(ambLat[i].Spellouts, mapping.Spellout)
			addedMissingSpellout = true
			ambLat[i].UnionPath(&lat)
		} else {
			// log.Println(mapping.Spellout, "Spellout found")
		}
		ambLat[i].BridgeMissingMorphemes()

		mappings[i] = mapping
	}

	m = &disambig.MDConfig{
		Mappings: mappings,
		Lattices: ambLat,
	}
	return m, addedMissingSpellout
}

func CombineLatticesCorpus(goldLats, ambLats []interface{}) ([]interface{}, int) {
	var (
		numLatticeNoGold int
	)
	prefix := log.Prefix()
	configs := make([]interface{}, 0, len(goldLats))
	for i, goldMap := range goldLats {
		ambLat := ambLats[i].(nlp.LatticeSentence)
		log.SetPrefix(fmt.Sprintf("%v graph# %v ", prefix, i))
		result, noGold := CombineToGoldMorph(goldMap.(nlp.LatticeSentence), ambLat)
		if noGold {
			numLatticeNoGold++
		}
		if result != nil {
			configs = append(configs, result)
		}
	}
	log.SetPrefix(prefix)
	return configs, numLatticeNoGold
}

func MDConfigOut(outModelFile string, b search.Interface, t transition.TransitionSystem) {
	log.Println("Configuration")
	log.Printf("Beam:             \t%s", b.Name())
	log.Printf("Transition System:\t%s", t.Name())
	log.Printf("Iterations:\t\t%d", Iterations)
	log.Printf("Beam Size:\t\t%d", BeamSize)
	log.Printf("Beam Concurrent:\t%v", ConcurrentBeam)
	log.Printf("Parameter Func:\t%v", paramFuncName)
	log.Printf("Use POP:\t\t%v", UsePOP)
	// log.Printf("Model file:\t\t%s", outModelFile)

	log.Println()
	log.Printf("Features File:\t%s", featuresFile)
	if !VerifyExists(featuresFile) {
		os.Exit(1)
	}
	log.Println()
	log.Println("Data")
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
	log.Printf("Out (disamb.) file:\t\t\t%s", outMap)
}

func MDTrainAndParse(cmd *commander.Command, args []string) {
	paramFunc, exists := nlp.MDParams[paramFuncName]
	if !exists {
		log.Fatalln("Param Func", paramFuncName, "does not exist")
	}
	var mdTrans transition.TransitionSystem
	if UseWB {
		mdTrans = &disambig.MDWBTrans{
			ParamFunc: paramFunc,
			UsePOP:    UsePOP,
		}
	} else {
		mdTrans = &disambig.MDTrans{
			ParamFunc: paramFunc,
			UsePOP:    UsePOP,
		}
	}
	disambig.UsePOP = UsePOP

	// arcSystem := &morph.Idle{morphArcSystem, IDLE}
	transitionSystem := transition.TransitionSystem(mdTrans)

	REQUIRED_FLAGS := []string{"it", "td", "tl", "in", "om", "f"}

	VerifyFlags(cmd, REQUIRED_FLAGS)
	// RegisterTypes()

	outModelFile := fmt.Sprintf("%s.b%d.i%d", modelFile, BeamSize, Iterations)
	confBeam := &search.Beam{}
	if !alignAverageParseOnly {
		confBeam.Align = AlignBeam
		confBeam.Averaged = AverageScores
	}

	MDConfigOut(outModelFile, confBeam, transitionSystem)

	if allOut {
		log.Println()
		// start processing - setup enumerations
		log.Println("Setup enumerations")
	}
	SetupMDEnum()
	if UseWB {
		mdTrans.(*disambig.MDWBTrans).POP = POP
		mdTrans.(*disambig.MDWBTrans).Transitions = ETrans
	} else {
		mdTrans.(*disambig.MDTrans).POP = POP
		mdTrans.(*disambig.MDTrans).Transitions = ETrans
	}
	mdTrans.AddDefaultOracle()
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

	if allOut {
		log.Println("Generating Gold Sequences For Training")
	}

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
	// goldDisLat = Limit(goldDisLat, 1000)

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
	// goldAmbLat = Limit(goldAmbLat, 1000)
	if allOut {
		log.Println("Combining train files into gold morph graphs with original lattices")
	}
	combined, missingGold := CombineLatticesCorpus(goldDisLat, goldAmbLat)

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
	goldSequences := TrainingSequences(combined, GetMDConfigAsLattices, GetMDConfigAsMappings)
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
	model := transitionmodel.NewAvgMatrixSparse(NumFeatures, formatters, false)

	conf := &disambig.MDConfig{
		ETokens: ETokens,
		POP:     POP,
	}

	beam := &search.Beam{
		TransFunc:            transitionSystem,
		FeatExtractor:        extractor,
		Base:                 conf,
		Size:                 BeamSize,
		ConcurrentExec:       ConcurrentBeam,
		Transitions:          ETrans,
		EstimatedTransitions: 1000, // chosen by random dice roll
	}

	if !alignAverageParseOnly {
		beam.Align = AlignBeam
		beam.Averaged = AverageScores
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

	var convCombined []interface{}
	var convDisLat []interface{}
	var convAmbLat []interface{}

	if len(inputGold) > 0 {
		log.Println("Reading test disambiguated lattice (for convergence testing) from", inputGold)
		lConvDis, lConvDisE := lattice.ReadFile(inputGold)
		if lConvDisE != nil {
			log.Println(lConvDisE)
			return
		}
		if allOut {
			log.Println("Convergence Test Gold Dis. Lat.:\tRead", len(lConvDis), "disambiguated lattices")
			log.Println("Convergence Test Gold Dis. Lat.:\tConverting lattice format to internal structure")
		}

		convDisLat = lattice.Lattice2SentenceCorpus(lConvDis, EWord, EPOS, EWPOS, EMorphProp, EMHost, EMSuffix)
		if allOut {
			log.Println("Reading test ambiguous lattices (for convergence testing) from", input)
		}

		lConvAmb, lConvAmbE := lattice.ReadFile(input)
		if lConvAmbE != nil {
			log.Println(lConvAmbE)
			return
		}
		// lAmb = lAmb[:NUM_SENTS]
		if allOut {
			log.Println("Read", len(lConvAmb), "ambiguous lattices from", input)
			log.Println("Converting lattice format to internal structure")
		}
		convAmbLat = lattice.Lattice2SentenceCorpus(lConvAmb, EWord, EPOS, EWPOS, EMorphProp, EMHost, EMSuffix)
		convCombined, _ = CombineLatticesCorpus(convDisLat, convAmbLat)

	}

	decodeTestBeam := &search.Beam{}
	*decodeTestBeam = *beam
	decodeTestBeam.Model = model
	decodeTestBeam.DecodeTest = true
	decodeTestBeam.ShortTempAgenda = true
	log.Println("Parse beam alignment:", AlignBeam)
	decodeTestBeam.Align = AlignBeam
	log.Println("Parse beam averaging:", AverageScores)
	decodeTestBeam.Averaged = AverageScores
	var evaluator perceptron.StopCondition
	if len(inputGold) > 0 {
		if allOut {
			log.Println("Setting convergence tester")
		}
		evaluator = MakeMorphEvalStopCondition(convAmbLat, convCombined, decodeTestBeam, perceptron.InstanceDecoder(deterministic), BeamSize)
	}
	_ = Train(goldSequences, Iterations, modelFile, model, perceptron.EarlyUpdateInstanceDecoder(beam), perceptron.InstanceDecoder(deterministic), evaluator)

	if allOut {
		log.Println("Done Training")
		// util.LogMemory()
		log.Println()
		// log.Println("Writing final model to", outModelFile)
		// WriteModel(model, outModelFile)
		// log.Println()
		log.Print("Parsing test")
	}
	if allOut {
		log.Println("Reading ambiguous lattices from", input)
	}

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

		_, missingGold = CombineLatticesCorpus(predDisLat, predAmbLat)

		if allOut {
			log.Println("Combined", len(predAmbLat), "graphs, with", missingGold, "missing at least one gold path in lattice")

			log.Println()
		}
	}
	beam.ShortTempAgenda = true
	log.Println("Parse beam alignment:", AlignBeam)
	beam.Align = AlignBeam
	log.Println("Parse beam averaging:", AverageScores)
	beam.Averaged = AverageScores
	beam.Model = model
	// mappings := Parse(predAmbLat, beam)
	//
	// /*	if allOut {
	// 		log.Println("Converting", len(parsedGraphs), "to conll")
	// 	}
	// */ // // // graphAsConll := conll.MorphGraph2ConllCorpus(parsedGraphs)
	// // // // if allOut {
	// // // // 	log.Println("Writing to output file")
	// // // // }
	// // // conll.WriteFile(outLat, graphAsConll)
	// // if allOut {
	// // 	log.Println("Wrote", len(graphAsConll), "in conll format to", outLat)
	//
	// // 	log.Println("Writing to segmentation file")
	// // }
	// // segmentation.WriteFile(outSeg, parsedGraphs)
	// // if allOut {
	// // 	log.Println("Wrote", len(parsedGraphs), "in segmentation format to", outSeg)
	//
	// // 	log.Println("Writing to gold segmentation file")
	// // }
	// // segmentation.WriteFile(tSeg, ToMorphGraphs(combined))
	//
	// if allOut {
	// 	log.Println("Writing to mapping file")
	// }
	// mapping.WriteFile(outMap, mappings)
	//
	// if allOut {
	// 	log.Println("Wrote", len(mappings), "in mapping format to", outMap)
	// }
}

func MdCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       MDTrainAndParse,
		UsageLine: "md <file options> [arguments]",
		Short:     "runs standalone morphological disambiguation training and parsing",
		Long: `
runs standalone morphological disambiguation training and parsing

	$ ./yap md -td <train disamb. lat> -tl <train amb. lat> -in <input lat> [-ing <input lat>] -om <out disamb> -f <feature file> [-p <param func>] [options]

`,
		Flag: *flag.NewFlagSet("md", flag.ExitOnError),
	}
	cmd.Flag.BoolVar(&ConcurrentBeam, "bconc", false, "Concurrent Beam")
	cmd.Flag.IntVar(&Iterations, "it", 1, "Number of Perceptron Iterations")
	cmd.Flag.IntVar(&BeamSize, "b", 4, "Beam Size")
	cmd.Flag.StringVar(&modelFile, "m", "model", "Prefix for model file ({m}.b{b}.i{it}.model)")

	cmd.Flag.StringVar(&tLatDis, "td", "", "Training Disambiguated Lattices File")
	cmd.Flag.StringVar(&tLatAmb, "tl", "", "Training Ambiguous Lattices File")
	cmd.Flag.StringVar(&input, "in", "", "Test Ambiguous Lattices File")
	cmd.Flag.StringVar(&inputGold, "ing", "", "Optional - Gold Test Lattices File (for infusion into test ambiguous)")
	cmd.Flag.StringVar(&outMap, "om", "", "Output Mapping File")
	cmd.Flag.StringVar(&featuresFile, "f", "", "Features Configuration File")
	cmd.Flag.StringVar(&paramFuncName, "p", "POS", "Param Func types: ["+nlp.AllParamFuncNames+"]")
	cmd.Flag.BoolVar(&AlignBeam, "align", false, "Use Beam Alignment")
	cmd.Flag.BoolVar(&AverageScores, "average", false, "Use Average Scoring")
	cmd.Flag.BoolVar(&alignAverageParseOnly, "parseonly", false, "Use Alignment & Average Scoring in parsing only")
	cmd.Flag.BoolVar(&UsePOP, "pop", false, "Add POP operation to MD")
	cmd.Flag.BoolVar(&UseWB, "wb", false, "Word Based MD")
	cmd.Flag.BoolVar(&search.AllOut, "showbeam", false, "Show candidates in beam")
	cmd.Flag.BoolVar(&search.ShowFeats, "showfeats", false, "Show features of candidates in beam")
	return cmd
}
