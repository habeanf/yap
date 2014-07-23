package md

import (
	"chukuparser/alg/perceptron"
	"chukuparser/alg/search"
	"chukuparser/alg/transition"
	transitionmodel "chukuparser/alg/transition/model"
	. "chukuparser/nlp/parser/disambig"

	"chukuparser/nlp/format/lattice"
	"chukuparser/nlp/format/mapping"

	nlp "chukuparser/nlp/types"
	"chukuparser/util"

	"encoding/gob"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func init() {
	gob.Register(&Serialization{})
}

var (
	allOut bool = true

	Iterations, BeamSize int
	ConcurrentBeam       bool
	NumFeatures          int

	// Global enumerations
	ETrans, EWord, EPOS, EWPOS, EMHost, EMSuffix *util.EnumSet
	ETokens                                      *util.EnumSet
	EMorphProp                                   *util.EnumSet

	tLatDis, tLatAmb string
	input            string
	inputGold        string
	outMap           string
	modelFile        string
	featuresFile     string
	paramFuncName    string

	REQUIRED_FLAGS []string = []string{"it", "td", "tl", "in", "om", "f"}
)

// An approximation of the number of different MD-X:Y:Z transitions
// Pre-allocating the enumeration saves frequent reallocation during training and parsing
const (
	APPROX_WORDS, APPROX_POS        = 100, 100
	WORDS_POS_FACTOR                = 5
	APPROX_MHOSTS, APPROX_MSUFFIXES = 128, 16
)

func SetupEnum() {
	EWord, EPOS, EWPOS = util.NewEnumSet(APPROX_WORDS), util.NewEnumSet(APPROX_POS), util.NewEnumSet(APPROX_WORDS*5)
	EMHost, EMSuffix = util.NewEnumSet(APPROX_MHOSTS), util.NewEnumSet(APPROX_MSUFFIXES)

	ETrans = util.NewEnumSet(10000)
	_, _ = ETrans.Add("NO") // dummy no action transition for zpar equivalence

	EMorphProp = util.NewEnumSet(130) // random guess of number of possible values
	ETokens = util.NewEnumSet(10000)
}

func SetupExtractor(setup *transition.FeatureSetup) *transition.GenericExtractor {
	extractor := &transition.GenericExtractor{
		EFeatures:  util.NewEnumSet(setup.NumFeatures()),
		Concurrent: false,
		EWord:      EWord,
		EPOS:       EPOS,
		EWPOS:      EWPOS,
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

func TrainingSequences(trainingSet []*MDConfig, transitionSystem transition.TransitionSystem, extractor perceptron.FeatureExtractor) []perceptron.DecodedInstance {
	instances := make([]perceptron.DecodedInstance, 0, len(trainingSet))

	for _, config := range trainingSet {
		sent := config.Lattices

		decoded := &perceptron.Decoded{sent, config}
		instances = append(instances, decoded)
	}
	return instances
}

func Train(trainingSet []perceptron.DecodedInstance, Iterations, BeamSize int, filename string, model perceptron.Model, transitionSystem transition.TransitionSystem, extractor perceptron.FeatureExtractor) *perceptron.LinearPerceptron {
	conf := &MDConfig{
		ETokens: ETokens,
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

	deterministic := &search.Deterministic{
		TransFunc:          transitionSystem,
		FeatExtractor:      extractor,
		ReturnModelValue:   false,
		ReturnSequence:     true,
		ShowConsiderations: false,
		Base:               conf,
		NoRecover:          false,
	}

	// varbeam := &VarBeam{beam}
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

	perceptron.Train(trainingSet)
	log.Println("TRAIN Total Time:", beam.DurTotal)

	return perceptron
}

func Parse(sents []nlp.LatticeSentence, BeamSize int, model transitionmodel.Interface, transitionSystem transition.TransitionSystem, extractor perceptron.FeatureExtractor) []nlp.Mappings {
	conf := &MDConfig{
		ETokens: ETokens,
	}

	beam := search.Beam{
		TransFunc:            transitionSystem,
		FeatExtractor:        extractor,
		Base:                 conf,
		Size:                 BeamSize,
		Model:                model,
		ConcurrentExec:       ConcurrentBeam,
		ShortTempAgenda:      true,
		Transitions:          ETrans,
		EstimatedTransitions: 1000,
	}

	// varbeam := &VarBeam{beam}

	mdSents := make([]nlp.Mappings, len(sents))
	for i, sent := range sents {
		// if i%100 == 0 {
		runtime.GC()
		for _, lat := range sent {
			lat.BridgeMissingMorphemes()
		}
		log.Println("Parsing sent", i)
		// }
		mapped, _ := beam.Parse(sent)
		mdSents[i] = mapped.(*MDConfig).Mappings
	}
	log.Println("PARSE Total Time:", beam.DurTotal)

	return mdSents
}

func CombineToGoldMorph(goldLat, ambLat nlp.LatticeSentence) (*MDConfig, bool) {
	var addedMissingSpellout bool
	// generate graph

	// generate morph. disambiguation (= mapping) and nodes
	mappings := make([]*nlp.Mapping, len(goldLat))
	for i, lat := range goldLat {
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
			ambLat[i].Spellouts = append(ambLat[i].Spellouts, mapping.Spellout)
			addedMissingSpellout = true
			ambLat[i].UnionPath(&lat)
		}
		ambLat[i].BridgeMissingMorphemes()

		mappings[i] = mapping
	}

	m := &MDConfig{
		Mappings: mappings,
		Lattices: ambLat,
	}
	return m, addedMissingSpellout
}

func CombineTrainingInputs(goldLats, ambLats []nlp.LatticeSentence) ([]*MDConfig, int) {
	var (
		numLatticeNoGold int
		noGold           bool
	)
	prefix := log.Prefix()
	configs := make([]*MDConfig, len(goldLats))
	for i, goldMap := range goldLats {
		ambLat := ambLats[i]
		log.SetPrefix(fmt.Sprintf("%v graph# %v ", prefix, i))
		configs[i], noGold = CombineToGoldMorph(goldMap, ambLat)
		if noGold {
			numLatticeNoGold++
		}
	}
	log.SetPrefix(prefix)
	return configs, numLatticeNoGold
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
		if f == nil || f.Value == nil || f.Value.String() == "" {
			log.Printf("Required flag %s not set", flag)
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
	log.Printf("Parameter Func:\t%v", paramFuncName)
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
		log.Printf("Test file  (disambig.  lattice):\t%s", input)
		if !VerifyExists(inputGold) {
			return
		}
	}
	log.Printf("Out (disamb.) file:\t\t\t%s", outMap)
}

func MD(cmd *commander.Command, args []string) {
	paramFunc, exists := MDParams[paramFuncName]
	if !exists {
		log.Fatalln("Param Func", paramFuncName, "does not exist")
	}
	mdTrans := &MDTrans{
		ParamFunc: paramFunc,
	}

	// arcSystem := &morph.Idle{morphArcSystem, IDLE}
	transitionSystem := transition.TransitionSystem(mdTrans)

	VerifyFlags(cmd)
	// RegisterTypes()

	outModelFile := fmt.Sprintf("%s.b%d.i%d", modelFile, BeamSize, Iterations)

	ConfigOut(outModelFile, &search.Beam{}, transitionSystem)

	if allOut {
		log.Println()
		// start processing - setup enumerations
		log.Println("Setup enumerations")
	}
	SetupEnum()
	mdTrans.Transitions = ETrans
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
	goldDisLat := lattice.Lattice2SentenceCorpus(lDis, EWord, EPOS, EWPOS, EMorphProp)

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
	goldAmbLat := lattice.Lattice2SentenceCorpus(lAmb, EWord, EPOS, EWPOS, EMorphProp)
	if allOut {
		log.Println("Combining train files into gold morph graphs with original lattices")
	}
	combined, missingGold := CombineTrainingInputs(goldDisLat, goldAmbLat)

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
	goldSequences := TrainingSequences(combined, transitionSystem, extractor)
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

	_ = Train(goldSequences, Iterations, BeamSize, modelFile, model, transitionSystem, extractor)
	if allOut {
		log.Println("Done Training")
		// util.LogMemory()
		log.Println()
		// log.Println("Writing final model to", outModelFile)
		// WriteModel(model, outModelFile)
		// log.Println()
		log.Print("Parsing test")

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
	predAmbLat := lattice.Lattice2SentenceCorpus(lAmb, EWord, EPOS, EWPOS, EMorphProp)

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

		predDisLat := lattice.Lattice2SentenceCorpus(lDis, EWord, EPOS, EWPOS, EMorphProp)

		if allOut {
			log.Println("Infusing test's gold disambiguation into ambiguous lattice")
		}

		_, missingGold = CombineTrainingInputs(predDisLat, predAmbLat)

		if allOut {
			log.Println("Combined", len(combined), "graphs, with", missingGold, "missing at least one gold path in lattice")

			log.Println()
		}
	}

	mappings := Parse(predAmbLat, BeamSize, transitionmodel.Interface(model), transitionSystem, extractor)

	/*	if allOut {
			log.Println("Converting", len(parsedGraphs), "to conll")
		}
	*/ // // // graphAsConll := conll.MorphGraph2ConllCorpus(parsedGraphs)
	// // // if allOut {
	// // // 	log.Println("Writing to output file")
	// // // }
	// // conll.WriteFile(outLat, graphAsConll)
	// if allOut {
	// 	log.Println("Wrote", len(graphAsConll), "in conll format to", outLat)

	// 	log.Println("Writing to segmentation file")
	// }
	// segmentation.WriteFile(outSeg, parsedGraphs)
	// if allOut {
	// 	log.Println("Wrote", len(parsedGraphs), "in segmentation format to", outSeg)

	// 	log.Println("Writing to gold segmentation file")
	// }
	// segmentation.WriteFile(tSeg, ToMorphGraphs(combined))

	if allOut {
		log.Println("Writing to mapping file")
	}
	mapping.WriteFile(outMap, mappings)

	if allOut {
		log.Println("Wrote", len(mappings), "in mapping format to", outMap)
	}
}

func MdCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       MD,
		UsageLine: "md <file options> [arguments]",
		Short:     "runs standalone morphological disambiguation training and parsing",
		Long: `
runs standalone morphological disambiguation training and parsing

	$ ./chukuparser md -td <train disamb. lat> -tl <train amb. lat> -in <input lat> [-ing <input lat>] -om <out disamb> -f <feature file> [-p <param func>] [options]

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
	paramFuncStrs := make([]string, 0, len(MDParams))
	for k, _ := range MDParams {
		paramFuncStrs = append(paramFuncStrs, k)
	}
	sort.Strings(paramFuncStrs)
	cmd.Flag.StringVar(&paramFuncName, "p", "POS", "Param Func types: ["+strings.Join(paramFuncStrs, ", ")+"]")
	return cmd
}

type Serialization struct {
	WeightModel                          *transitionmodel.AvgMatrixSparseSerialized
	EWord, EPOS, EWPOS, EMHost, EMSuffix *util.EnumSet
	ETrans                               *util.EnumSet
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
