package Transition

import (
	"chukuparser/NLP"
)

var TEST_SENT TaggedSentence = TaggedSentence{
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

func GetTestDepGraph() NLP.LabeledDependencyGraph {
	var (
		nodes []NLP.DepNode  = make([]NLP.DepNode, len(rawNodes))
		arcs  []*BasicDepArc = make([]*BasicDepArc, len(rawArcs))
	)
	for i, rawNode := range rawNodes {
		nodes[i] = NLP.DepNode(&rawNode)
	}
	for i, rawArc := range rawArcs {
		// make sure to get use a heap pointer with it's own copy
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
	// Set up configuration:
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
