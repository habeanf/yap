package Morph

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/NLP/Parser/Dependency"

	G "chukuparser/Algorithm/Graph"
	Transition "chukuparser/Algorithm/Transition"
	T "chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
	"log"
	"runtime"
	"sort"
	"testing"
)

// sent: HELIM MZHIBIM .
// lattice:
// 0	1	H		_	REL		REL		_								1
// 0	1	H		_	DEF		DEF		_								1
// 0	2	HELIM	_	VB		VB		gen=M|num=S|per=3|tense=PAST	1
// 1	2	ELIM	_	NN		NN		gen=M|num=P						1
// 2	3	MZHIBIM	_	BN		BN		gen=M|num=P|per=A				2
// 2	3	MZHIBIM	_	VB		VB		gen=M|num=P|per=A|tense=BEINONI	2
// 3	4	yyDOT	_	yyDOT	yyDOT	_								3

var TEST_LATTICE NLP.LatticeSentence = NLP.LatticeSentence{
	{"HELIM",
		[]*NLP.Morpheme{
			&NLP.Morpheme{G.BasicDirectedEdge{0, 0, 1}, "H", "REL", "REL",
				nil, 1},
			&NLP.Morpheme{G.BasicDirectedEdge{1, 0, 1}, "H", "DEF", "DEF",
				nil, 1},
			&NLP.Morpheme{G.BasicDirectedEdge{2, 0, 2}, "HELIM", "VB", "VB",
				map[string]string{"gen": "M", "num": "S", "per": "3", "tense": "PAST"}, 1},
			&NLP.Morpheme{G.BasicDirectedEdge{3, 1, 2}, "ELIM", "NN", "NN",
				map[string]string{"gen": "M", "num": "P"}, 1},
		},
		nil,
	},
	{"MZHIBIM",
		[]*NLP.Morpheme{
			&NLP.Morpheme{G.BasicDirectedEdge{0, 2, 3}, "MZHIBIM", "BN", "BN",
				map[string]string{"gen": "M", "num": "P", "per": "A"}, 2},
			&NLP.Morpheme{G.BasicDirectedEdge{1, 2, 3}, "MZHIBIM", "VB", "VB",
				map[string]string{"gen": "M", "num": "P", "P": "A", "tense": "BEINONI"}, 2},
		},
		nil,
	},
	{"yyDOT",
		[]*NLP.Morpheme{
			&NLP.Morpheme{G.BasicDirectedEdge{0, 3, 4}, "yyDOT", "yyDOT", "yyDOT",
				nil, 3},
		},
		nil,
	},
}

// SENT: HELIM MZHIBIM .
// gold dep:
// 0	1	H		_	DEF		DEF		_					2	def		_	_
// 1	2	ELIM	_	NN		NN		gen=M|num=P			3	subj	_	_
// 2	3	MZHIBIM	_	BN		BN		gen=M|num=P|per=A	0	prd		_	_
// 3	4	yyDOT	_	yyDOT	yyDOT	_					3	punct	_	_

// mapping:
// 1 HELIM		H:ELIM
// 2 MZHIBIM	MZHIBIM
// 3 yyDOT		yyDOT

var TEST_GRAPH *BasicMorphGraph = &BasicMorphGraph{
	T.BasicDepGraph{
		[]NLP.DepNode{
			&NLP.Morpheme{G.BasicDirectedEdge{0, 0, 0}, "ROOT", "ROOT", "ROOT",
				nil, 0},
			&NLP.Morpheme{G.BasicDirectedEdge{1, 0, 1}, "H", "DEF", "DEF",
				nil, 1},
			&NLP.Morpheme{G.BasicDirectedEdge{3, 1, 2}, "ELIM", "NN", "NN",
				map[string]string{"gen": "M", "num": "P"}, 1},
			&NLP.Morpheme{G.BasicDirectedEdge{0, 2, 3}, "MZHIBIM", "BN", "BN",
				map[string]string{"gen": "M", "num": "P", "per": "A"}, 2},
			&NLP.Morpheme{G.BasicDirectedEdge{0, 3, 4}, "yyDOT", "yyDOT", "yyDOT",
				nil, 3},
		},
		[]*T.BasicDepArc{
			&T.BasicDepArc{2, "def", 1},
			&T.BasicDepArc{3, "subj", 2},
			&T.BasicDepArc{0, "prd", 3},
			&T.BasicDepArc{3, "punct", 4},
		},
	},
	[]*NLP.Mapping{
		// &NLP.Mapping{"ROOT", []*NLP.Morpheme{}},
		&NLP.Mapping{"HELIM", []*NLP.Morpheme{
			&NLP.Morpheme{G.BasicDirectedEdge{1, 0, 1}, "H", "DEF", "DEF",
				nil, 1},
			&NLP.Morpheme{G.BasicDirectedEdge{3, 1, 2}, "ELIM", "NN", "NN",
				map[string]string{"gen": "M", "num": "P"}, 1},
		}},
		&NLP.Mapping{"MZHIBIM", []*NLP.Morpheme{
			&NLP.Morpheme{G.BasicDirectedEdge{0, 2, 3}, "MZHIBIM", "BN", "BN",
				map[string]string{"gen": "M", "num": "P", "per": "A"}, 2},
		}},
		&NLP.Mapping{"yyDOT", []*NLP.Morpheme{
			&NLP.Morpheme{G.BasicDirectedEdge{0, 3, 4}, "yyDOT", "yyDOT", "yyDOT",
				nil, 3},
		}},
	},
	nil,
}

var TEST_MORPH_TRANSITIONS []string = []string{
	"MD-1", "SH", "LA-def", "SH", "MD-0", "LA-subj", "RA-prd", "MD-0", "RA-punct",
}

var TEST_RELATIONS []string = []string{
	"advmod", "amod", "appos", "aux",
	"cc", "ccomp", "comp", "complmn",
	"compound", "conj", "cop", "def",
	"dep", "det", "detmod", "gen",
	"ghd", "gobj", "hd", "mod",
	"mwe", "neg", "nn", "null",
	"num", "number", "obj", "parataxis",
	"pcomp", "pobj", "posspmod", "prd",
	"prep", "prepmod", "punct", "qaux",
	"rcmod", "rel", "relcomp", "subj",
	"tmod", "xcomp",
}

//ALL RICH FEATURES
var TEST_RICH_FEATURES []string = []string{
	"S0|w|p", "S0|w", "S0|p", "N0|w|p",
	"N0|w", "N0|p", "N1|w|p", "N1|w",
	"N1|p", "N2|w|p", "N2|w", "N2|p",
	"S0|w|p+N0|w|p", "S0|w|p+N0|w",
	"S0|w+N0|w|p", "S0|w|p+N0|p",
	"S0|p+N0|w|p", "S0|w+N0|w",
	"S0|p+N0|p", "N0|p+N1|p",
	"N0|p+N1|p+N2|p", "S0|p+N0|p+N1|p",
	"S0h|p+S0|p+N0|p", "S0|p+S0l|p+N0|p",
	"S0|p+S0r|p+N0|p", "S0|p+N0|p+N0l|p",
	"S0|w|d", "S0|p|d", "N0|w|d", "N0|p|d",
	"S0|w+N0|w|d", "S0|p+N0|p|d",
	"S0|w|vr", "S0|p|vr", "S0|w|vl", "S0|p|vl", "N0|w|vl", "N0|p|vl",
	"S0h|w", "S0h|p", "S0|l", "S0l|w",
	"S0l|p", "S0l|l", "S0r|w", "S0r|p",
	"S0r|l", "N0l|w", "N0l|p", "N0l|l",
	"S0h2|w", "S0h2|p", "S0h|l", "S0l2|w",
	"S0l2|p", "S0l2|l", "S0r2|w", "S0r2|p",
	"S0r2|l", "N0l2|w", "N0l2|p", "N0l2|l",
	"S0|p+S0l|p+S0l2|p", "S0|p+S0r|p+S0r2|p",
	"S0|p+S0h|p+S0h2|p", "N0|p+N0l|p+N0l2|p",
	"S0|w|sr", "S0|p|sr", "S0|w|sl", "S0|p|sl",
	"N0|w|sl", "N0|p|sl"}

func TestOracle(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("Testing Oracle")
	mconf := new(MorphConfiguration)
	mconf.Init(TEST_LATTICE)
	conf := Transition.Configuration(mconf)
	arcmorph := new(ArcEagerMorph)
	arcmorph.AddDefaultOracle()
	trans := Transition.TransitionSystem(arcmorph)
	trans.Oracle().SetGold(TEST_GRAPH)

	goldTrans := TEST_MORPH_TRANSITIONS
	for !conf.Terminal() {
		oracle := trans.Oracle()
		transition := oracle.Transition(conf)
		// log.Println("Chose transition:", transition)
		if string(transition) != goldTrans[0] {
			t.Error("Gold is:", goldTrans[0], "got", transition)
			return
		}
		conf = trans.Transition(conf, transition)
		goldTrans = goldTrans[1:]
	}
	log.Println("Done testing Oracle")
	// log.Println("\n", conf.GetSequence().String())
}

func TestDeterministic(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("Testing Deterministic")
	runtime.GOMAXPROCS(runtime.NumCPU())
	extractor := new(T.GenericExtractor)
	// verify load
	for _, feature := range TEST_RICH_FEATURES {
		if err := extractor.LoadFeature(feature); err != nil {
			t.Error("Failed to load feature", err.Error())
			t.FailNow()
		}
	}
	arcSystem := &ArcEagerMorph{}
	arcSystem.Relations = TEST_RELATIONS
	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)
	deterministic := &T.Deterministic{transitionSystem, extractor, true, true, false, &MorphConfiguration{}}
	decoder := Perceptron.EarlyUpdateInstanceDecoder(deterministic)
	updater := new(Perceptron.AveragedStrategy)

	goldInstances := []Perceptron.DecodedInstance{
		&Perceptron.Decoded{Perceptron.Instance(TEST_LATTICE), TEST_GRAPH}}

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init()
	goldModel := Dependency.ParameterModel(&T.PerceptronModel{perceptron})
	graph := TEST_GRAPH
	graph.Lattice = TEST_LATTICE

	_, goldParams := deterministic.ParseOracle(graph, nil, goldModel)
	goldSequence := goldParams.(*T.ParseResultParameters).Sequence

	// train with increasing iterations
	convergenceIterations := []int{1, 2, 8, 16, 32}
	convergenceSharedSequence := make([]int, 0, len(convergenceIterations))
	for _, iterations := range convergenceIterations {
		perceptron.Iterations = iterations
		// perceptron.Log = true
		perceptron.Init()

		// deterministic.ShowConsiderations = true
		perceptron.Train(goldInstances)

		model := Dependency.ParameterModel(&T.PerceptronModel{perceptron})
		// deterministic.ShowConsiderations = true
		_, params := deterministic.Parse(TEST_LATTICE, nil, model)
		seq := params.(*T.ParseResultParameters).Sequence
		// log.Println(seq)
		sharedSteps := goldSequence.SharedTransitions(seq)
		convergenceSharedSequence = append(convergenceSharedSequence, sharedSteps)
	}

	// verify convergence
	if !sort.IntsAreSorted(convergenceSharedSequence) || convergenceSharedSequence[0] == convergenceSharedSequence[len(convergenceSharedSequence)-1] {
		t.Error("Model not converging, shared sequences lengths:", convergenceSharedSequence)
	}
	log.Println("Done Testing Deterministic")
}

func TestSimpleBeam(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("Testing Simple Beam")
	runtime.GOMAXPROCS(runtime.NumCPU())
	// runtime.GOMAXPROCS(1)
	extractor := new(T.GenericExtractor)
	// verify load
	for _, feature := range TEST_RICH_FEATURES {
		if err := extractor.LoadFeature(feature); err != nil {
			t.Error("Failed to load feature", err.Error())
			t.FailNow()
		}
	}
	arcSystem := &ArcEagerMorph{}
	arcSystem.Relations = TEST_RELATIONS
	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)

	conf := &MorphConfiguration{}

	beam := &T.Beam{
		TransFunc:     transitionSystem,
		FeatExtractor: extractor,
		Base:          conf,
		NumRelations:  len(arcSystem.Relations),
	}

	decoder := Perceptron.EarlyUpdateInstanceDecoder(beam)
	updater := new(Perceptron.AveragedStrategy)

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init()
	graph := TEST_GRAPH
	graph.Lattice = TEST_LATTICE

	// get gold parse
	goldModel := Dependency.ParameterModel(&T.PerceptronModel{perceptron})
	deterministic := &T.Deterministic{transitionSystem, extractor, true, true, false, conf}
	_, goldParams := deterministic.ParseOracle(graph, nil, goldModel)
	goldSequence := goldParams.(*T.ParseResultParameters).Sequence

	goldInstances := []Perceptron.DecodedInstance{
		&Perceptron.Decoded{Perceptron.Instance(TEST_LATTICE), goldSequence[0]}}

	// train with increasing iterations
	beam.ConcurrentExec = true
	beam.ReturnSequence = true

	convergenceIterations := []int{1, 4, 16}
	beamSizes := []int{1, 4, 16, 64}
	for _, beamSize := range beamSizes {
		beam.Size = beamSize

		convergenceSharedSequence := make([]int, 0, len(convergenceIterations))
		for _, iterations := range convergenceIterations {
			perceptron.Iterations = iterations
			// perceptron.Log = true
			perceptron.Init()

			// deterministic.ShowConsiderations = true
			perceptron.Train(goldInstances)

			model := Dependency.ParameterModel(&T.PerceptronModel{perceptron})
			beam.ReturnModelValue = false
			_, params := beam.Parse(TEST_LATTICE, nil, model)
			sharedSteps := 0
			if params != nil {
				seq := params.(*T.ParseResultParameters).Sequence
				sharedSteps = goldSequence.SharedTransitions(seq)
			}
			convergenceSharedSequence = append(convergenceSharedSequence, sharedSteps)
		}
		if len(convergenceSharedSequence) != len(convergenceIterations) {
			t.Error("Not enough examples in shared sequence samples")
		}
		// verify convergence
		log.Println("Shared Sequence For Beam", beamSize, convergenceSharedSequence)
		if !sort.IntsAreSorted(convergenceSharedSequence) || convergenceSharedSequence[len(convergenceSharedSequence)-1] == 0 {
			t.Error("Model not converging, shared sequences lengths:", convergenceSharedSequence)
		}
	}
	log.Println("Done Testing Simple Beam")
}
