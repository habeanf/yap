package morphparse

import (
	"chukuparser/algorithm/perceptron"
	"chukuparser/algorithm/transition"
	transitionmodel "chukuparser/algorithm/transition/model"
	"chukuparser/nlp/format/conll"
	"chukuparser/nlp/format/lattice"
	"chukuparser/nlp/format/segmentation"
	"chukuparser/nlp/parser/dependency"
	. "chukuparser/nlp/parser/dependency/transition"
	"chukuparser/nlp/parser/dependency/transition/morph"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"
	"chukuparser/util/conf"

	// "encoding/gob"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var (
	allOut bool = true

	Iterations, BeamSize int
	ConcurrentBeam       bool
	NumFeatures          int

	// Global enumerations
	ERel, ETrans, EWord, EPOS, EWPOS *Util.EnumSet

	// Enumeration offsets of transitions
	SH, RE, PR, IDLE, LA, RA, MD Transition.Transition

	tConll, tLatDis, tLatAmb string
	tSeg                     string
	input                    string
	outLat, outSeg           string
	modelFile                string
	featuresFile             string
	labelsFile               string

	REQUIRED_FLAGS []string = []string{"it", "tc", "td", "tl", "in", "oc", "os", "ots", "f", "l"}
)

func SetupRelationEnum(labels []string) {
	if ERel != nil {
		return
	}
	ERel = Util.NewEnumSet(len(labels) + 1)
	ERel.Add(nlp.DepRel(nlp.ROOT_LABEL))
	for _, label := range labels {
		ERel.Add(nlp.DepRel(label))
	}
	ERel.Frozen = true
}

// An approximation of the number of different MD-X:Y:Z transitions
// Pre-allocating the enumeration saves frequent reallocation during training and parsing
const (
	APPROX_MORPH_TRANSITIONS = 100
	APPROX_WORDS, APPROX_POS = 100, 100
	WORDS_POS_FACTOR         = 5
)

func SetupMorphTransEnum(relations []string) {
	ETrans = Util.NewEnumSet((len(relations)+1)*2 + 2 + APPROX_MORPH_TRANSITIONS)
	_, _ = ETrans.Add("NO") // dummy for 0 action
	iSH, _ := ETrans.Add("SH")
	iRE, _ := ETrans.Add("RE")
	_, _ = ETrans.Add("AL") // dummy action transition for zpar equivalence
	_, _ = ETrans.Add("AR") // dummy action transition for zpar equivalence
	iPR, _ := ETrans.Add("PR")
	// iIDLE, _ := ETrans.Add("IDLE")
	SH = Transition.Transition(iSH)
	RE = Transition.Transition(iRE)
	PR = Transition.Transition(iPR)
	// IDLE = Transition.Transition(iIDLE)
	// LA = IDLE + 1
	LA = PR + 1
	ETrans.Add("LA-" + string(nlp.ROOT_LABEL))
	for _, transition := range relations {
		ETrans.Add("LA-" + string(transition))
	}
	RA = Transition.Transition(ETrans.Len())
	ETrans.Add("RA-" + string(nlp.ROOT_LABEL))
	for _, transition := range relations {
		ETrans.Add("RA-" + string(transition))
	}
	MD = Transition.Transition(ETrans.Len())
}

func SetupEnum(relations []string) {
	SetupRelationEnum(relations)
	SetupMorphTransEnum(relations)
	EWord, EPOS, EWPOS = Util.NewEnumSet(APPROX_WORDS), Util.NewEnumSet(APPROX_POS), Util.NewEnumSet(APPROX_WORDS*5)
}

func SetupExtractor(features []string) *GenericExtractor {
	extractor := &GenericExtractor{
		EFeatures:  Util.NewEnumSet(len(features)),
		Concurrent: false,
		EWord:      EWord,
		EPOS:       EPOS,
		EWPOS:      EWPOS,
		ERel:       ERel,
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

func TrainingSequences(trainingSet []*Morph.BasicMorphGraph, transitionSystem Transition.TransitionSystem, extractor Perceptron.FeatureExtractor) []Perceptron.DecodedInstance {
	// verify feature load

	mconf := &Morph.MorphConfiguration{
		SimpleConfiguration: SimpleConfiguration{
			EWord:  EWord,
			EPOS:   EPOS,
			EWPOS:  EWPOS,
			ERel:   ERel,
			ETrans: ETrans,
		},
	}

	deterministic := &Deterministic{
		TransFunc:          transitionSystem,
		FeatExtractor:      extractor,
		ReturnModelValue:   false,
		ReturnSequence:     true,
		ShowConsiderations: false,
		Base:               mconf,
		// NoRecover:          true,
	}

	decoder := Perceptron.EarlyUpdateInstanceDecoder(deterministic)
	updater := new(transitionmodel.AveragedModelStrategy)

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	model := transitionmodel.NewAvgMatrixSparse(NumFeatures, nil)

	tempModel := Dependency.TransitionParameterModel(&PerceptronModel{model})
	perceptron.Init(model)

	instances := make([]Perceptron.DecodedInstance, 0, len(trainingSet))
	var failedTraining int
	for i, graph := range trainingSet {
		if i%100 == 0 {
			log.Println("At line", i)
		}
		sent := graph.Lattice

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
	conf := &Morph.MorphConfiguration{
		SimpleConfiguration: SimpleConfiguration{
			EWord:  EWord,
			EPOS:   EPOS,
			EWPOS:  EWPOS,
			ERel:   ERel,
			ETrans: ETrans,
		},
	}

	beam := &Beam{
		TransFunc:      transitionSystem,
		FeatExtractor:  extractor,
		Base:           conf,
		NumRelations:   ERel.Len(),
		Size:           BeamSize,
		ConcurrentExec: ConcurrentBeam,
	}

	// varbeam := &VarBeam{beam}
	decoder := Perceptron.EarlyUpdateInstanceDecoder(beam)
	updater := new(transitionmodel.AveragedModelStrategy)

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

	return perceptron
}

func Parse(sents []nlp.LatticeSentence, BeamSize int, model Dependency.TransitionParameterModel, transitionSystem Transition.TransitionSystem, extractor Perceptron.FeatureExtractor) []nlp.MorphDependencyGraph {
	conf := &Morph.MorphConfiguration{
		SimpleConfiguration: SimpleConfiguration{
			EWord:  EWord,
			EPOS:   EPOS,
			EWPOS:  EWPOS,
			ERel:   ERel,
			ETrans: ETrans,
		},
	}

	beam := Beam{
		TransFunc:       transitionSystem,
		FeatExtractor:   extractor,
		Base:            conf,
		Size:            BeamSize,
		NumRelations:    ERel.Len(),
		Model:           model,
		ConcurrentExec:  ConcurrentBeam,
		ShortTempAgenda: true}

	// varbeam := &VarBeam{beam}

	parsedGraphs := make([]nlp.MorphDependencyGraph, len(sents))
	for i, sent := range sents {
		// if i%100 == 0 {
		runtime.GC()
		log.Println("Parsing sent", i)
		// }
		graph, _ := beam.Parse(sent, nil, model)
		labeled := graph.(nlp.MorphDependencyGraph)
		parsedGraphs[i] = labeled
	}
	log.Println("PARSE Total Time:", beam.DurTotal)

	return parsedGraphs
}

func CombineTrainingInputs(graphs []nlp.LabeledDependencyGraph, goldLats, ambLats []nlp.LatticeSentence) ([]*Morph.BasicMorphGraph, int) {
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

func ConfigOut(outModelFile string) {
	log.Println("Configuration")
	// log.Printf("Beam:             \tVariable Length")
	log.Printf("Beam:             \tStatic Length")
	// log.Printf("Transition System:\tIDLE + Morph + ArcEager")
	log.Printf("Transition System:\tMorph + ArcEager")
	log.Printf("Iterations:\t\t%d", Iterations)
	log.Printf("Beam Size:\t\t%d", BeamSize)
	log.Printf("Beam Concurrent:\t%v", ConcurrentBeam)
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
	log.Printf("Out (disamb.) file:\t\t\t%s", outLat)
	log.Printf("Out (segmt.) file:\t\t\t%s", outSeg)
	log.Printf("Out Train (segmt.) file:\t\t%s", tSeg)
}

func MorphTrainAndParse(cmd *commander.Command, args []string) {
	VerifyFlags(cmd)
	// RegisterTypes()

	outModelFile := fmt.Sprintf("%s.b%d.i%d", modelFile, BeamSize, Iterations)

	ConfigOut(outModelFile)

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
	s, e := Conll.ReadFile(tConll)
	if e != nil {
		log.Println(e)
		return
	}
	if allOut {
		log.Println("Conll:\tRead", len(s), "sentences")
		log.Println("Conll:\tConverting from conll to internal structure")
	}
	goldConll := Conll.Conll2GraphCorpus(s, EWord, EPOS, EWPOS, ERel)

	if allOut {
		log.Println("Dis. Lat.:\tReading training disambiguated lattices from", tLatDis)
	}
	lDis, lDisE := Lattice.ReadFile(tLatDis)
	if lDisE != nil {
		log.Println(lDisE)
		return
	}
	if allOut {
		log.Println("Dis. Lat.:\tRead", len(lDis), "disambiguated lattices")
		log.Println("Dis. Lat.:\tConverting lattice format to internal structure")
	}
	goldDisLat := Lattice.Lattice2SentenceCorpus(lDis, EWord, EPOS, EWPOS)

	if allOut {
		log.Println("Amb. Lat:\tReading ambiguous lattices from", tLatAmb)
	}
	lAmb, lAmbE := Lattice.ReadFile(tLatAmb)
	if lAmbE != nil {
		log.Println(lAmbE)
		return
	}
	if allOut {
		log.Println("Amb. Lat:\tRead", len(lAmb), "ambiguous lattices")
		log.Println("Amb. Lat:\tConverting lattice format to internal structure")
	}
	goldAmbLat := Lattice.Lattice2SentenceCorpus(lAmb, EWord, EPOS, EWPOS)
	if allOut {
		log.Println("Combining train files into gold morph graphs with original lattices")
	}
	combined, missingGold := CombineTrainingInputs(goldConll, goldDisLat, goldAmbLat)

	if allOut {
		log.Println("Combined", len(combined), "graphs, with", missingGold, "missing at least one gold path in lattice")

		log.Println()

	}

	morphArcSystem := &Morph.ArcEagerMorph{
		ArcEager: ArcEager{
			ArcStandard: ArcStandard{
				SHIFT:       SH,
				LEFT:        LA,
				RIGHT:       RA,
				Relations:   ERel,
				Transitions: ETrans,
			},
			REDUCE:  RE,
			POPROOT: PR},
		MD: MD,
	}
	morphArcSystem.AddDefaultOracle()

	// arcSystem := &Morph.Idle{morphArcSystem, IDLE}
	transitionSystem := Transition.TransitionSystem(morphArcSystem)

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
		// Util.LogMemory()
		log.Println("Training", Iterations, "iteration(s)")
	}
	formatters := make([]Util.Format, len(extractor.FeatureTemplates))
	for i, formatter := range extractor.FeatureTemplates {
		formatters[i] = formatter
	}
	model := transitionmodel.NewAvgMatrixSparse(NumFeatures, formatters)
	_ = Train(goldSequences, Iterations, BeamSize, modelFile, model, transitionSystem, extractor)
	if allOut {
		log.Println("Done Training")
		// Util.LogMemory()
		log.Println()
		// log.Println("Writing final model to", outModelFile)
		// WriteModel(model, outModelFile)
		// log.Println()
		log.Print("Parsing test")

		log.Println("Reading ambiguous lattices from", input)
	}
	lAmb, lAmbE = Lattice.ReadFile(input)
	if lAmbE != nil {
		log.Println(lAmbE)
		return
	}
	// lAmb = lAmb[:NUM_SENTS]
	if allOut {
		log.Println("Read", len(lAmb), "ambiguous lattices from", input)
		log.Println("Converting lattice format to internal structure")
	}
	predAmbLat := Lattice.Lattice2SentenceCorpus(lAmb, EWord, EPOS, EWPOS)

	parsedGraphs := Parse(predAmbLat, BeamSize, Dependency.TransitionParameterModel(&PerceptronModel{model}), transitionSystem, extractor)

	if allOut {
		log.Println("Converting", len(parsedGraphs), "to conll")
	}
	graphAsConll := Conll.MorphGraph2ConllCorpus(parsedGraphs)
	if allOut {
		log.Println("Writing to output file")
	}
	Conll.WriteFile(outLat, graphAsConll)
	if allOut {
		log.Println("Wrote", len(graphAsConll), "in conll format to", outLat)

		log.Println("Writing to segmentation file")
	}
	Segmentation.WriteFile(outSeg, parsedGraphs)
	if allOut {
		log.Println("Wrote", len(parsedGraphs), "in segmentation format to", outSeg)

		log.Println("Writing to gold segmentation file")
	}
	Segmentation.WriteFile(tSeg, ToMorphGraphs(combined))
	if allOut {
		log.Println("Wrote", len(combined), "in segmentation format to", tSeg)
	}
}

func ToMorphGraphs(graphs []*Morph.BasicMorphGraph) []nlp.MorphDependencyGraph {
	morphs := make([]nlp.MorphDependencyGraph, len(graphs))
	for i, g := range graphs {
		morphs[i] = nlp.MorphDependencyGraph(g)
	}
	return morphs
}

func MorphCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       MorphTrainAndParse,
		UsageLine: "morph <file options> [arguments]",
		Short:     "runs morpho-syntactic training and parsing",
		Long: `
runs morpho-syntactic training and parsing

	$ ./chukuparser morph -tc <conll> -td <train disamb. lat> -tl <train amb. lat> -in <input lat> -oc <out lat> -os <out seg> -ots <out train seg> [options]

`,
		Flag: *flag.NewFlagSet("morph", flag.ExitOnError),
	}
	cmd.Flag.BoolVar(&ConcurrentBeam, "bconc", false, "Concurrent Beam")
	cmd.Flag.IntVar(&Iterations, "it", 1, "Number of Perceptron Iterations")
	cmd.Flag.IntVar(&BeamSize, "b", 4, "Beam Size")
	cmd.Flag.StringVar(&modelFile, "m", "model", "Prefix for model file ({m}.b{b}.i{it}.model)")

	cmd.Flag.StringVar(&tConll, "tc", "", "Training Conll File")
	cmd.Flag.StringVar(&tLatDis, "td", "", "Training Disambiguated Lattices File")
	cmd.Flag.StringVar(&tLatAmb, "tl", "", "Training Ambiguous Lattices File")
	cmd.Flag.StringVar(&input, "in", "", "Test Ambiguous Lattices File")
	cmd.Flag.StringVar(&outLat, "oc", "", "Output Conll File")
	cmd.Flag.StringVar(&outSeg, "os", "", "Output Segmentation File")
	cmd.Flag.StringVar(&tSeg, "ots", "", "Output Training Segmentation File")
	cmd.Flag.StringVar(&featuresFile, "f", "", "Features Configuration File")
	cmd.Flag.StringVar(&labelsFile, "l", "", "Dependency Labels Configuration File")
	return cmd
}
