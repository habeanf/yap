package app

import (
	"chukuparser/alg/perceptron"
	"chukuparser/alg/search"
	"chukuparser/alg/transition"

	"chukuparser/alg/transition/model"
	"chukuparser/nlp/parser/disambig"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"

	"chukuparser/nlp/parser/dependency/transition/morph"

	"encoding/gob"
	"log"
	"os"
	"runtime"
	"time"
	// "strings"

	"github.com/gonuts/commander"
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
	ETokens                                            *util.EnumSet
	EMorphProp                                         *util.EnumSet

	// enumeration offsets of transitions
	SH, RE, PR, LA, RA, IDLE, MD transition.Transition

	// file names
	tConll           string
	tLatDis, tLatAmb string
	tSeg             string
	input            string
	inputGold        string
	outLat, outSeg   string
	outMap           string
	outConll         string
	modelFile        string
	featuresFile     string
	labelsFile       string
)

// An approximation of the number of different MD-X:Y:Z transitions
// Pre-allocating the enumeration saves frequent reallocation during training and parsing
const (
	APPROX_MORPH_TRANSITIONS        = 100
	APPROX_WORDS, APPROX_POS        = 100, 100
	WORDS_POS_FACTOR                = 5
	APPROX_MHOSTS, APPROX_MSUFFIXES = 128, 16
)

type Serialization struct {
	WeightModel                          *model.AvgMatrixSparseSerialized
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

func SetupMorphTransEnum(relations []string) {
	ETrans = util.NewEnumSet((len(relations)+1)*2 + 2 + APPROX_MORPH_TRANSITIONS)
	_, _ = ETrans.Add("NO") // dummy for 0 action
	iSH, _ := ETrans.Add("SH")
	iRE, _ := ETrans.Add("RE")
	_, _ = ETrans.Add("AL") // dummy action transition for zpar equivalence
	_, _ = ETrans.Add("AR") // dummy action transition for zpar equivalence
	iPR, _ := ETrans.Add("PR")
	// iIDLE, _ := ETrans.Add("IDLE")
	SH = transition.Transition(iSH)
	RE = transition.Transition(iRE)
	PR = transition.Transition(iPR)
	// IDLE = transition.Transition(iIDLE)
	// LA = IDLE + 1
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
	log.Println("ETrans Len is", ETrans.Len())
	MD = transition.Transition(ETrans.Len())
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

func VerifyFlags(cmd *commander.Command, required []string) {
	for _, flag := range required {
		f := cmd.Flag.Lookup(flag)
		if f.Value.String() == "" {
			log.Printf("Required flag %s not set", f.Name)
			cmd.Usage()
			os.Exit(1)
		}
	}
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

	NumFeatures = setup.NumFeatures()
	return extractor
}

type InstanceFunc func(interface{}) util.Equaler
type GoldFunc func(interface{}) util.Equaler

func TrainingSequences(trainingSet []interface{}, instFunc InstanceFunc, goldFunc GoldFunc) []perceptron.DecodedInstance {
	instances := make([]perceptron.DecodedInstance, 0, len(trainingSet))

	for i, instance := range trainingSet {
		log.Println("At training", i)

		decoded := &perceptron.Decoded{instFunc(instance), goldFunc(instance)}
		instances = append(instances, decoded)
	}
	return instances
}

func Train(trainingSet []perceptron.DecodedInstance, Iterations int, filename string, paramModel perceptron.Model, decoder perceptron.EarlyUpdateInstanceDecoder, goldDecoder perceptron.InstanceDecoder) *perceptron.LinearPerceptron {
	updater := new(model.AveragedModelStrategy)

	perceptron := &perceptron.LinearPerceptron{
		Decoder:     decoder,
		GoldDecoder: goldDecoder,
		Updater:     updater,
		Tempfile:    filename,
		TempLines:   1000}

	perceptron.Iterations = Iterations
	perceptron.Init(paramModel)
	// perceptron.TempLoad("model.b64.i1")
	perceptron.Log = true
	// beam.Log = true
	startTime := time.Now()
	perceptron.Train(trainingSet)
	if allOut {
		trainTime := time.Since(startTime)
		log.Println("TRAIN Total Time:", trainTime)
	}
	return perceptron
}

type Parser interface {
	Parse(search.Problem) (transition.Configuration, interface{})
}

func Parse(instances []interface{}, parser Parser) []interface{} {
	// runtime.GOMAXPROCS(1)
	// Search.AllOut = true
	startTime := time.Now()

	parsed := make([]interface{}, len(instances))
	for i, instance := range instances {
		if i%5 == 0 {
			runtime.GC()
		}
		log.Println("Parsing instance", i) //, "len", len(sent.Tokens()))
		// }
		result, _ := parser.Parse(instance)
		parsed[i] = result
	}
	if allOut {
		parseTime := time.Since(startTime)
		log.Println("PARSE Total Time:", parseTime)
	}
	return parsed
}

func GetMDConfigAsLattices(instance interface{}) util.Equaler {
	return instance.(*disambig.MDConfig).Lattices
}

func GetMDConfigAsMappings(instance interface{}) util.Equaler {
	return instance.(*disambig.MDConfig).Mappings
}

func GetMorphGraphAsLattices(instance interface{}) util.Equaler {
	return instance.(*morph.BasicMorphGraph).Lattice
}

func GetMorphGraph(instance interface{}) util.Equaler {
	return instance.(*morph.BasicMorphGraph)
}

func GetAsTaggedSentence(instance interface{}) util.Equaler {
	return instance.(nlp.LabeledDependencyGraph).TaggedSentence()
}

func GetAsLabeledDepGraph(instance interface{}) util.Equaler {
	return instance.(nlp.LabeledDependencyGraph)
}
