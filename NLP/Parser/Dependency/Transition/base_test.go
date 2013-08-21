package Transition

import (
	NLP "chukuparser/NLP/Types"
)

var TEST_SENT NLP.TaggedSentence = NLP.BasicTaggedSentence{
	{"Economic", "NN"},
	{"news", "NN"},
	{"had", "VB"},
	{"little", "ADJ"},
	{"effect", "NN"},
	{"on", "NN"},
	{"financial", "NN"},
	{"markets", "NN"},
	{".", "yyDOT"}}

var rawNodes []TaggedDepNode = []TaggedDepNode{
	{0, ROOT_TOKEN, ROOT_TOKEN},
	{1, "Economic", "NN"},
	{2, "news", "NN"},
	{3, "had", "VB"},
	{4, "little", "ADJ"},
	{5, "effect", "NN"},
	{6, "on", "NN"},
	{7, "financial", "NN"},
	{8, "markets", "NN"},
	{9, ".", "yyDOT"}}

var rawArcs []BasicDepArc = []BasicDepArc{
	{2, "ATT", 1},
	{3, "SBJ", 2},
	{5, "ATT", 4},
	{8, "ATT", 7},
	{6, "PC", 8},
	{5, "ATT", 6},
	{3, "OBJ", 5},
	{3, "PU", 9},
	{0, "PRED", 3}}

var TEST_RELATIONS []string = []string{"ATT", "SBJ", "PC", "OBJ", "PU", "PRED"}

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

func GetTestDepGraph() NLP.LabeledDependencyGraph {
	var (
		nodes []NLP.DepNode  = make([]NLP.DepNode, len(rawNodes))
		arcs  []*BasicDepArc = make([]*BasicDepArc, len(rawArcs))
	)
	for i, rawNode := range rawNodes {
		node := new(TaggedDepNode)
		*node = rawNode
		nodes[i] = NLP.DepNode(node)
	}
	for i, rawArc := range rawArcs {
		// make sure to get a heap pointer with it's own copy
		// otherwise &rawArc will be constant
		newArcPtr := new(BasicDepArc)
		*newArcPtr = rawArc
		arcs[i] = newArcPtr
	}
	return NLP.LabeledDependencyGraph(&BasicDepGraph{nodes, arcs})
}

func GetTestConfiguration() *SimpleConfiguration {
	conf := new(SimpleConfiguration)
	conf.Init(TEST_SENT)
	// [ROOT Economic news had little effect on financial markets .]
	//   0      1      2    3    4      5    6      7       8     9
	// Setup configuration:
	// C=(	[ROOT,had,effect], [.], A)
	// A={	(ROOT,	PRED,	had)
	// 		(had,	OBJ,	effect)
	// 		(effect,ATT,	little)
	//		(effect,ATT,	on)}

	// S=[ROOT,had,effect]
	// stack should already have ROOT
	if peekVal, peekExists := conf.Stack().Peek(); !peekExists || peekVal != 0 {
		panic("Initialized configuration should have root as head of stack")
	}
	conf.Stack().Push(3)
	conf.Stack().Push(5)

	// B=[.]
	conf.Queue().Clear()
	conf.Queue().Push(9)

	// A = {...}
	conf.Arcs().Add(&BasicDepArc{0, "PRED", 3})
	conf.Arcs().Add(&BasicDepArc{3, "OBJ", 5})
	conf.Arcs().Add(&BasicDepArc{5, "ATT", 4})
	conf.Arcs().Add(&BasicDepArc{5, "ATT", 6})

	return conf
}
