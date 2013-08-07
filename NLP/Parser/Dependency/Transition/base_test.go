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
