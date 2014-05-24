package md

import (
	// "chukuparser/algorithm/featurevector"
	"chukuparser/algorithm/perceptron"
	BeamSearch "chukuparser/algorithm/search"
	"chukuparser/algorithm/transition"
	transitionmodel "chukuparser/algorithm/transition/model"
	"chukuparser/nlp/format/conll"
	"chukuparser/nlp/format/lattice"
	"chukuparser/nlp/format/segmentation"
	. "chukuparser/nlp/parser/disambig"
	// . "chukuparser/nlp/parser/dependency/transition"
	"chukuparser/nlp/parser/dependency/transition/morph"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"
	"chukuparser/util/conf"

	"encoding/gob"
	"fmt"
	"log"
	"os"
	"runtime"
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
	ETrans, EWord, EPOS, EWPOS *util.EnumSet
	EMorphProp                 *util.EnumSet

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
	APPROX_MORPH_TRANSITIONS = 100
	APPROX_WORDS, APPROX_POS = 100, 100
	WORDS_POS_FACTOR         = 5
)

func SetupEnum() {
	EWord, EPOS, EWPOS = util.NewEnumSet(APPROX_WORDS), util.NewEnumSet(APPROX_POS), util.NewEnumSet(APPROX_WORDS*5)
	EMorphProp = util.NewEnumSet(130) // random guess of number of possible values
}

func SetupExtractor(features []string) *transition.GenericExtractor {
	extractor := &transition.GenericExtractor{
		EFeatures:  util.NewEnumSet(len(features)),
		Concurrent: false,
		EWord:      EWord,
		EPOS:       EPOS,
		EWPOS:      EWPOS,
	}
	extractor.Init()
	for _, feature := range features {
		featurePair := strings.Split(feature, ",")
		if err := extractor.LoadFeature(featurePair[0], featurePair[1]); err != nil {
			log.Fatalln("Failed to load feature", err.Error())
		}
	}
	NumFeatures = len(features)
	return extractor
}

func TrainingSequences(trainingSet []*morph.BasicMorphGraph, transitionSystem transition.TransitionSystem, extractor perceptron.FeatureExtractor) []perceptron.DecodedInstance {
	instances := make([]perceptron.DecodedInstance, 0, len(trainingSet))

	for _, graph := range trainingSet {
		sent := graph.Lattice

		decoded := &perceptron.Decoded{sent, graph}
		instances = append(instances, decoded)
	}
	return instances
}

func Train(trainingSet []perceptron.DecodedInstance, Iterations, BeamSize int, filename string, model perceptron.Model, transitionSystem transition.TransitionSystem, extractor perceptron.FeatureExtractor) *perceptron.LinearPerceptron {
	conf := &MDConfig{}

	beam := &Beam{
		TransFunc:      transitionSystem,
		FeatExtractor:  extractor,
		Base:           conf,
		Size:           BeamSize,
		ConcurrentExec: ConcurrentBeam,
	}

	// varbeam := &VarBeam{beam}
	decoder := perceptron.EarlyUpdateInstanceDecoder(beam)
	updater := new(transitionmodel.AveragedModelStrategy)

	perceptron := &perceptron.LinearPerceptron{
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

func CombineTrainingInputs(graphs []nlp.LabeledDependencyGraph, goldLats, ambLats []nlp.LatticeSentence) ([]*morph.BasicMorphGraph, int) {
	if len(graphs) != len(goldLats) || len(graphs) != len(ambLats) {
		panic(fmt.Sprintf("Got mismatched training slice inputs (graphs, gold lattices, ambiguous lattices):", len(graphs), len(goldLats), len(ambLats)))
	}
	morphGraphs := make([]*morph.BasicMorphGraph, len(graphs))
	var (
		numLatticeNoGold int
		noGold           bool
	)
	prefix := log.Prefix()
	for i, goldGraph := range graphs {
		goldLat := goldLats[i]
		ambLat := ambLats[i]
		log.SetPrefix(fmt.Sprintf("%v graph# %v ", prefix, i))
		morphGraphs[i], noGold = morph.CombineToGoldMorph(goldGraph, goldLat, ambLat)
		if noGold {
			numLatticeNoGold++
		}
	}
	log.SetPrefix(prefix)
	return morphGraphs, numLatticeNoGold
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
	mdTrans := &MDTrans{}
	mdTrans.AddDefaultOracle()

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

	if allOut {
		log.Println()
		log.Println("Loading features")
	}
	features, err := conf.ReadFile(featuresFile)
	if err != nil {
		log.Println("Failed reading feature configuration file:", featuresFile)
		log.Fatalln(err)
	}
	extractor := SetupExtractor(features.Values)

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
	goldConll := conll.Conll2GraphCorpus(s, EWord, EPOS, EWPOS, nil, nil, nil)

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
	combined, missingGold := CombineTrainingInputs(goldConll, goldDisLat, goldAmbLat)

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
	segmentation.WriteFile(tSeg, ToMorphGraphs(combined))
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
