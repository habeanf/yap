package md

import (
	// "chukuparser/algorithm/featurevector"
	"chukuparser/algorithm/perceptron"
	BeamSearch "chukuparser/algorithm/search"
	"chukuparser/algorithm/transition"
	transitionmodel "chukuparser/algorithm/transition/model"
	"chukuparser/nlp/format/lattice"
	// "chukuparser/nlp/format/segmentation"
	. "chukuparser/nlp/parser/disambig"
	// . "chukuparser/nlp/parser/dependency/transition"
	"chukuparser/nlp/parser/dependency/transition/morph"
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

	tConll, tLatDis, tLatAmb string
	tSeg                     string
	input                    string
	outLat, outSeg           string
	modelFile                string
	featuresFile             string
	labelsFile               string

	REQUIRED_FLAGS []string = []string{"it", "td", "tl", "in", "oc", "os", "ots", "f"}
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

	beam := &Beam{
		TransFunc:     transitionSystem,
		FeatExtractor: extractor,
		Base:          conf,
		Size:          BeamSize,
		// ConcurrentExec: ConcurrentBeam,
		ConcurrentExec: false,
		Transitions:    ETrans,
	}

	deterministic := &Deterministic{
		TransFunc:          transitionSystem,
		FeatExtractor:      extractor,
		ReturnModelValue:   false,
		ReturnSequence:     true,
		ShowConsiderations: false,
		Base:               conf,
		NoRecover:          true,
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
	conf := &MDConfig{}

	beam := Beam{
		TransFunc:       transitionSystem,
		FeatExtractor:   extractor,
		Base:            conf,
		Size:            BeamSize,
		Model:           model,
		ConcurrentExec:  ConcurrentBeam,
		ShortTempAgenda: true,
		Transitions:     ETrans,
	}

	// varbeam := &VarBeam{beam}

	mdSents := make([]nlp.Mappings, len(sents))
	for i, sent := range sents {
		// if i%100 == 0 {
		runtime.GC()
		log.Println("Parsing sent", i)
		// }
		mapped, _ := beam.Parse(sent)
		mdSents[i] = mapped
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
		}

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

func ConfigOut(outModelFile string, b BeamSearch.Interface, t transition.TransitionSystem) {
	log.Println("Configuration")
	// log.Printf("Beam:             \tVariable Length")
	log.Printf("Beam:             \t%s", b.Name())
	// log.Printf("Transition System:\tIDLE + Morph + ArcEager")
	log.Printf("Transition System:\t%s", t.Name())
	log.Printf("Iterations:\t\t%d", Iterations)
	log.Printf("Beam Size:\t\t%d", BeamSize)
	log.Printf("Beam Concurrent:\t%v", ConcurrentBeam)
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
	log.Printf("Out (disamb.) file:\t\t\t%s", outLat)
	log.Printf("Out (segmt.) file:\t\t\t%s", outSeg)
	log.Printf("Out Train (segmt.) file:\t\t%s", tSeg)
}

func MD(cmd *commander.Command, args []string) {
	mdTrans := &MDTrans{
		ParamFunc: Full,
	}

	// arcSystem := &morph.Idle{morphArcSystem, IDLE}
	transitionSystem := transition.TransitionSystem(mdTrans)

	VerifyFlags(cmd)
	// RegisterTypes()

	outModelFile := fmt.Sprintf("%s.b%d.i%d", modelFile, BeamSize, Iterations)

	ConfigOut(outModelFile, &Beam{}, transitionSystem)

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
	model := transitionmodel.NewAvgMatrixSparse(NumFeatures, formatters, true)
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

	_ = Parse(predAmbLat, BeamSize, transitionmodel.Interface(model), transitionSystem, extractor)

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
		log.Println("Wrote", len(combined), "in segmentation format to", tSeg)
	}
}

func ToMorphGraphs(graphs []*morph.BasicMorphGraph) []nlp.MorphDependencyGraph {
	morphs := make([]nlp.MorphDependencyGraph, len(graphs))
	for i, g := range graphs {
		morphs[i] = nlp.MorphDependencyGraph(g)
	}
	return morphs
}

func MdCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       MD,
		UsageLine: "md <file options> [arguments]",
		Short:     "runs standalone morphological disambiguation training and parsing",
		Long: `
runs standalone morphological disambiguation training and parsing

	$ ./chukuparser md -td <train disamb. lat> -tl <train amb. lat> -in <input lat> -oc <out disamb> -os <out seg> -ots <out train seg> -f <feature file> [options]

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
	cmd.Flag.StringVar(&outLat, "oc", "", "Output Disambiguated Lattices File")
	cmd.Flag.StringVar(&outSeg, "os", "", "Output Segmentation File")
	cmd.Flag.StringVar(&tSeg, "ots", "", "Output Training Segmentation File")
	cmd.Flag.StringVar(&featuresFile, "f", "", "Features Configuration File")
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
