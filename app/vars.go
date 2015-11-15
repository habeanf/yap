package app

import (
	"yap/alg/perceptron"
	"yap/alg/search"
	"yap/alg/transition"

	"yap/alg/transition/model"
	// dep "yap/nlp/parser/dependency/transition"
	"yap/eval"
	"yap/nlp/format/conll"
	"yap/nlp/format/mapping"
	"yap/nlp/parser/disambig"
	"yap/nlp/parser/joint"
	nlp "yap/nlp/types"
	"yap/util"

	dep "yap/nlp/parser/dependency/transition"
	"yap/nlp/parser/dependency/transition/morph"

	"encoding/gob"
	"fmt"
	"log"
	"os"
	// "runtime"
	"time"
	// "strings"
	"runtime"
	"runtime/debug"

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
	UsePOP               bool

	// global enumerations
	ERel, ETrans, EWord, EPOS, EWPOS, EMHost, EMSuffix *util.EnumSet
	ETokens                                            *util.EnumSet
	EMorphProp                                         *util.EnumSet

	// enumeration offsets of transitions
	SH, RE, PR, LA, RA, IDLE, POP, MD transition.Transition

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

	AlignBeam             bool
	AverageScores         bool
	alignAverageParseOnly bool

	arcSystemStr string
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
	EMorphProp                           *util.EnumSet
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
	_, _ = ETrans.Add("IDLE") // dummy no action transition for zpar equivalence
	iSH, _ := ETrans.Add("SH")
	iRE, _ := ETrans.Add("RE")
	_, _ = ETrans.Add("AL") // dummy action transition for zpar equivalence
	_, _ = ETrans.Add("AR") // dummy action transition for zpar equivalence
	iPR, _ := ETrans.Add("PR")
	SH = transition.ConstTransition(iSH)
	RE = transition.ConstTransition(iRE)
	PR = transition.ConstTransition(iPR)
	LA = transition.ConstTransition(iPR + 1)
	ETrans.Add("LA-" + string(nlp.ROOT_LABEL))
	for _, transition := range relations {
		ETrans.Add("LA-" + string(transition))
	}
	RA = transition.ConstTransition(ETrans.Len())
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
	SH = transition.ConstTransition(iSH)
	RE = transition.ConstTransition(iRE)
	PR = transition.ConstTransition(iPR)
	// IDLE = transition.Transition(iIDLE)
	// LA = IDLE + 1
	LA = transition.ConstTransition(iPR + 1)
	ETrans.Add("LA-" + string(nlp.ROOT_LABEL))
	for _, transition := range relations {
		ETrans.Add("LA-" + string(transition))
	}
	RA = transition.ConstTransition(ETrans.Len())
	ETrans.Add("RA-" + string(nlp.ROOT_LABEL))
	for _, transition := range relations {
		ETrans.Add("RA-" + string(transition))
	}
	log.Println("ETrans Len is", ETrans.Len())
	MD = transition.ConstTransition(ETrans.Len())
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

func SetupExtractor(setup *transition.FeatureSetup, transTypes []byte) *transition.GenericExtractor {
	extractor := &transition.GenericExtractor{
		EFeatures:  util.NewEnumSet(setup.NumFeatures()),
		Concurrent: false,
		EWord:      EWord,
		EPOS:       EPOS,
		EWPOS:      EWPOS,
		ERel:       ERel,
		EMHost:     EMHost,
		EMSuffix:   EMSuffix,
		EMorphProp: EMorphProp,
		POPTrans:   POP,
		// Log:        true,
	}
	if transTypes == nil {
		extractor.Init()
	} else {
		extractor.InitTypes(transTypes)
	}
	extractor.LoadFeatureSetup(setup)

	NumFeatures = setup.NumFeatures()
	return extractor
}

type InstanceFunc func(interface{}) util.Equaler
type GoldFunc func(interface{}) util.Equaler

func Limit(instances []interface{}, limit int) []interface{} {
	if len(instances) > limit {
		return instances[:limit]
	}
	return instances
}

func TrainingSequences(trainingSet []interface{}, instFunc InstanceFunc, goldFunc GoldFunc) []perceptron.DecodedInstance {
	instances := make([]perceptron.DecodedInstance, 0, len(trainingSet))

	for _, instance := range trainingSet {
		// log.Println("At training", i)

		decoded := &perceptron.Decoded{instFunc(instance), goldFunc(instance)}
		instances = append(instances, decoded)
	}
	return instances
}

// Assumes sorted inputs of equal length
func DepEval(test, gold interface{}) *eval.Result {
	testConf, testOk := test.(*dep.SimpleConfiguration)
	goldGraph, goldOk := gold.(*dep.BasicDepGraph)
	// log.Println(testMorph.GetSequence())
	// log.Println(goldMorph.GetSequence())
	if !testOk {
		panic("Test argument should be MDConfig")
	}
	if !goldOk {
		panic("Gold argument should be nlp.Mappings")
	}
	testArcs := testConf.Arcs().(*dep.ArcSetSimple).Arcs
	goldArcs := goldGraph.Arcs
	retval := &eval.Result{ // retval is LAS
		Other: &eval.Result{}, // Other is UAS evaluation
	}
	// log.Println("Test is:")
	// log.Println(testArcs)
	// log.Println("Gold is:")
	// log.Println(goldArcs)
	var unlabeledAttached, labeledAttached bool
	for _, curTestArc := range testArcs {
		unlabeledAttached, labeledAttached = false, false
		for _, curGoldArc := range goldArcs {
			if curTestArc.GetHead() == curGoldArc.GetHead() &&
				curTestArc.GetModifier() == curGoldArc.GetModifier() {
				unlabeledAttached = true
				retval.Other.(*eval.Result).TP += 1
				if curTestArc.GetRelation() == curGoldArc.GetRelation() {
					labeledAttached = true
					retval.TP += 1
				}
				break
			}
		}
		if !labeledAttached {
			retval.FP += 1
		}
		if !unlabeledAttached {
			retval.Other.(*eval.Result).FP += 1
		}
	}
	for _, curTestArc := range testArcs {
		unlabeledAttached, labeledAttached = false, false
		for _, curGoldArc := range goldArcs {
			if curTestArc.GetHead() == curGoldArc.GetHead() &&
				curTestArc.GetModifier() == curGoldArc.GetModifier() {
				unlabeledAttached = true
				if curTestArc.GetRelation() == curGoldArc.GetRelation() {
					labeledAttached = true
				}
				break
			}
		}
		if !labeledAttached {
			retval.TN += 1
		}
		if !unlabeledAttached {
			retval.Other.(*eval.Result).TN += 1
		}
	}
	return retval
}

// Assumes sorted inputs of equal length
func MorphEval(test, gold interface{}, metric string) *eval.Result {
	testMorph, testOk := test.(*disambig.MDConfig)
	goldMappings, goldOk := gold.(nlp.Mappings)
	// log.Println(testMorph.GetSequence())
	// log.Println(goldMorph.GetSequence())
	if !testOk {
		panic("Test argument should be MDConfig")
	}
	if !goldOk {
		panic("Gold argument should be nlp.Mappings")
	}
	testMappings := testMorph.Mappings
	retval := &eval.Result{}
	// log.Println("Test is:")
	// log.Println(testMappings)
	// log.Println("Gold is:")
	// log.Println(goldMappings)
	for i, testMapping := range testMappings {
		goldMapping := goldMappings[i]
		// if testMapping.Token != goldMapping.Token {
		// 	panic(fmt.Sprintf("Mappings #%v are not equal: %v %v", i, testMapping.Token, goldMapping.Token))
		// }
		testSpellout := testMapping.Spellout
		goldSpellout := goldMapping.Spellout
		TP, TN, FP, FN := testSpellout.Compare(goldSpellout, metric)
		retval.TP += TP
		retval.TN += TN
		retval.FP += FP
		retval.FN += FN
	}
	return retval
}

func Train(trainingSet []perceptron.DecodedInstance, Iterations int, filename string, paramModel perceptron.Model, decoder perceptron.EarlyUpdateInstanceDecoder, goldDecoder perceptron.InstanceDecoder, converge perceptron.StopCondition) *perceptron.LinearPerceptron {
	updater := new(model.AveragedModelStrategy)

	perceptron := &perceptron.LinearPerceptron{
		Decoder:     decoder,
		GoldDecoder: goldDecoder,
		Updater:     updater,
		Continue:    converge,
		Tempfile:    filename,
		TempLines:   100}

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

	prevGC := debug.SetGCPercent(-1)
	parsed := make([]interface{}, len(instances))
	for i, instance := range instances {
		if i%100 == 0 {
			debug.SetGCPercent(100)
			runtime.GC()
			debug.SetGCPercent(-1)
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
	debug.SetGCPercent(prevGC)
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

func GetJointMDConfig(instance interface{}) util.Equaler {
	return &instance.(*joint.JointConfig).MDConfig
}

func GetInstances(instances []interface{}, getFunc InstanceFunc) []interface{} {
	retval := make([]interface{}, len(instances))
	for i, val := range instances {
		retval[i] = getFunc(val)
	}
	return retval
}

func MakeMorphEvalStopCondition(instances []interface{}, goldInstances []interface{}, parser Parser, goldDecoder perceptron.InstanceDecoder, beamSize int) perceptron.StopCondition {
	var (
		equalIterations int
		prevResult      float64
	)
	return func(curIteration, iterations, generations int) bool {
		// log.Println("Eval starting for iteration", curIteration)
		var total = &eval.Total{
			Results: make([]*eval.Result, 0, len(instances)),
		}
		var posonlytotal = &eval.Total{
			Results: make([]*eval.Result, 0, len(instances)),
		}
		// Don't test before initial run
		if curIteration == 0 {
			return true
		}
		var curResult float64
		var curPosResult float64
		// TODO: fix this leaky abstraction :(
		// log.Println("Temp integration using", generations)
		parser.(*search.Beam).IntegrationGeneration = generations
		parsed := Parse(instances, parser)
		goldInstances := TrainingSequences(goldInstances, GetMDConfigAsLattices, GetMDConfigAsMappings)
		log.Println("START Evaluation")
		if len(goldInstances) != len(instances) {
			panic("Evaluation instance lengths are different")
		}
		for i, instance := range parsed {
			// log.Println("Evaluating", i)
			goldInstance := goldInstances[i]
			if goldInstance != nil {
				result := MorphEval(instance, goldInstance.Decoded(), "Form_POS_Prop")
				posresult := MorphEval(instance, goldInstance.Decoded(), "Form_POS")
				// log.Println("Correct: ", result.TP)
				total.Add(result)
				posonlytotal.Add(posresult)
			}
		}
		curResult = total.F1()
		curPosResult = posonlytotal.F1()
		// Break out of edge case where result remains the same
		if curResult == prevResult {
			equalIterations += 1
		}
		retval := curResult < prevResult || equalIterations > 2
		// retval := curIteration >= iterations
		log.Println("Result (F1): ", curResult, "Exact:", total.Exact, "TruePos:", total.TP, "in", total.Population, "POS F1:", curPosResult)
		if retval {
			log.Println("Stopping")
		} else {
			log.Println("Continuing")
		}
		prevResult = curResult
		mapping.WriteFile(fmt.Sprintf("interm.i%v.b%v.%v", curIteration, beamSize, outMap), parsed)
		return !retval
	}
}

func MakeDepEvalStopCondition(instances []interface{}, goldInstances []interface{}, parser Parser, goldDecoder perceptron.InstanceDecoder, beamSize int) perceptron.StopCondition {
	var (
		equalIterations     int
		prevResult          float64
		continuousDecreases int
	)
	return func(curIteration, iterations, generations int) bool {
		// log.Println("Eval starting for iteration", curIteration)
		var total = &eval.Total{
			Results: make([]*eval.Result, 0, len(instances)),
		}
		// Don't test before initial run
		if curIteration == 0 {
			return true
		}
		var curResult float64
		// TODO: fix this leaky abstraction :(
		// log.Println("Temp integration using", generations)
		parser.(*search.Beam).IntegrationGeneration = generations
		oldparseOut := parseOut
		parseOut = true
		parsed := Parse(instances, parser)
		parseOut = oldparseOut
		goldInstances := TrainingSequences(goldInstances, GetAsTaggedSentence, GetAsLabeledDepGraph)
		log.Println("START Evaluation")
		if len(goldInstances) != len(instances) {
			panic("Evaluation instance lengths are different")
		}
		for i, instance := range parsed {
			// log.Println("Evaluating", i)
			goldInstance := goldInstances[i]
			if goldInstance != nil {
				result := DepEval(instance, goldInstance.Decoded())
				// log.Println("Correct: ", result.TP)
				total.Add(result)
			}
		}
		curResult = total.Precision()
		// Break out of edge case where result remains the same
		if curResult == prevResult {
			equalIterations += 1
		}
		retval := (continuousDecreases > 0 && curResult < prevResult) || equalIterations > 2
		// retval := curIteration >= iterations
		log.Println("Result (LAS precision): ", curResult, "Exact:", total.Exact, "TruePos:", total.TP, "in", total.Population)
		if retval {
			log.Println("Stopping")
		} else {
			log.Println("Continuing")
		}
		if curResult < prevResult {
			continuousDecreases += 1
		} else {
			continuousDecreases = 0
		}
		prevResult = curResult
		graphs := conll.Graph2ConllCorpus(parsed, EMHost, EMSuffix)
		conll.WriteFile(fmt.Sprintf("interm.i%v.b%v.%v", curIteration, beamSize, outConll), graphs)
		return !retval
	}
}
