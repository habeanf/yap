package Morph

import (
	G "chukuparser/Algorithm/Graph"
	Transition "chukuparser/Algorithm/Transition"
	T "chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
	"log"
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

var TEST_GRAPH NLP.MorphDependencyGraph = &BasicMorphGraph{
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
		&NLP.Mapping{"ROOT", []*NLP.Morpheme{}},
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
}

var TEST_MORPH_TRANSITIONS []string = []string{
	"MD-1", "SH", "LA-def", "SH", "MD-0", "LA-subj", "RA-prd", "MD-0", "RA-punct",
}

func TestMorphConfig(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
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
		log.Println("Chose transition:", transition)
		if string(transition) != goldTrans[0] {
			t.Error("Gold is:", goldTrans[0])
			return
		}
		conf = trans.Transition(conf, transition)
		goldTrans = goldTrans[1:]
	}
	log.Println("\n", conf.GetSequence().String())
}

func TestOracle(t *testing.T) {

}
