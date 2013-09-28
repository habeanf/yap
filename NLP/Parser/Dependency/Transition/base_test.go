package Transition

import (
	AbstractTransition "chukuparser/Algorithm/Transition"
	NLP "chukuparser/NLP/Types"
	"chukuparser/Util"
)

var rawTestSent NLP.BasicETaggedSentence = NLP.BasicETaggedSentence{
	{TaggedToken: NLP.TaggedToken{"Economic", "NN"}},
	{TaggedToken: NLP.TaggedToken{"news", "NN"}},
	{TaggedToken: NLP.TaggedToken{"had", "VB"}},
	{TaggedToken: NLP.TaggedToken{"little", "ADJ"}},
	{TaggedToken: NLP.TaggedToken{"effect", "NN"}},
	{TaggedToken: NLP.TaggedToken{"on", "NN"}},
	{TaggedToken: NLP.TaggedToken{"financial", "NN"}},
	{TaggedToken: NLP.TaggedToken{"markets", "NN"}},
	{TaggedToken: NLP.TaggedToken{".", "yyDOT"}}}

var TEST_SENT NLP.TaggedSentence

var rawNodes []TaggedDepNode = []TaggedDepNode{
	{Id: 0, RawToken: NLP.ROOT_TOKEN, RawPOS: NLP.ROOT_TOKEN},
	{Id: 1, RawToken: "Economic", RawPOS: "NN"},
	{Id: 2, RawToken: "news", RawPOS: "NN"},
	{Id: 3, RawToken: "had", RawPOS: "VB"},
	{Id: 4, RawToken: "little", RawPOS: "ADJ"},
	{Id: 5, RawToken: "effect", RawPOS: "NN"},
	{Id: 6, RawToken: "on", RawPOS: "NN"},
	{Id: 7, RawToken: "financial", RawPOS: "NN"},
	{Id: 8, RawToken: "markets", RawPOS: "NN"},
	{Id: 9, RawToken: ".", RawPOS: "yyDOT"}}

var rawArcs []BasicDepArc = []BasicDepArc{
	{Head: 2, RawRelation: NLP.DepRel("ATT"), Modifier: 1},
	{Head: 3, RawRelation: NLP.DepRel("SBJ"), Modifier: 2},
	{Head: 5, RawRelation: NLP.DepRel("ATT"), Modifier: 4},
	{Head: 8, RawRelation: NLP.DepRel("ATT"), Modifier: 7},
	{Head: 6, RawRelation: NLP.DepRel("PC"), Modifier: 8},
	{Head: 5, RawRelation: NLP.DepRel("ATT"), Modifier: 6},
	{Head: 3, RawRelation: NLP.DepRel("OBJ"), Modifier: 5},
	{Head: 3, RawRelation: NLP.DepRel("PU"), Modifier: 9},
	{Head: 0, RawRelation: NLP.DepRel(NLP.ROOT_LABEL), Modifier: 3}}

var (
	TEST_RELATIONS      []NLP.DepRel = []NLP.DepRel{"ATT", "SBJ", "PC", "OBJ", "PU", "PRED", NLP.ROOT_LABEL}
	TRANSITIONS_ENUM    *Util.EnumSet
	TEST_ENUM_RELATIONS *Util.EnumSet
	EWord, EPOS, EWPOS  *Util.EnumSet
	SH, RE, PR, LA, RA  AbstractTransition.Transition
)

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

func SetupRelationEnum() {
	if TEST_ENUM_RELATIONS != nil {
		return
	}
	TEST_ENUM_RELATIONS = Util.NewEnumSet(len(TEST_RELATIONS))
	for _, label := range TEST_RELATIONS {
		TEST_ENUM_RELATIONS.Add(label)
	}
}

func SetupSentEnum() {
	EWord, EPOS, EWPOS =
		Util.NewEnumSet(len(rawNodes)),
		Util.NewEnumSet(5), // 4 POS + ROOT
		Util.NewEnumSet(len(rawNodes))
	var (
		// val   int
		node  *TaggedDepNode
		arc   *BasicDepArc
		token *NLP.EnumTaggedToken
	)
	EWord.Add(NLP.ROOT_TOKEN)
	EPOS.Add(NLP.ROOT_TOKEN)
	EWPOS.Add([2]string{NLP.ROOT_TOKEN, NLP.ROOT_TOKEN})
	for i, _ := range rawNodes {
		node = &rawNodes[i]
		node.Token, _ = EWord.Add(node.RawToken)
		node.POS, _ = EPOS.Add(node.RawPOS)
		node.TokenPOS, _ = EWPOS.Add([2]string{node.RawToken, node.RawPOS})
	}
	for i, _ := range rawArcs {
		arc = &rawArcs[i]
		arc.Relation, _ = TEST_ENUM_RELATIONS.Add(arc.RawRelation)
	}
	for i, _ := range rawTestSent {
		token = &rawTestSent[i]
		token.EToken, _ = EWord.Add(token.Token)
		token.EPOS, _ = EPOS.Add(token.POS)
		token.ETPOS, _ = EWPOS.Add([2]string{token.Token, token.POS})
	}
	TEST_SENT = NLP.TaggedSentence(rawTestSent)
}

func SetupTestEnum() {
	SetupRelationEnum()
	SetupSentEnum()
}

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
	SetupTestEnum()
	SetupEagerTransEnum() // default trans is eager
	conf := &SimpleConfiguration{
		EWord:  EWord,
		EPOS:   EPOS,
		EWPOS:  EWPOS,
		ERel:   TEST_ENUM_RELATIONS,
		ETrans: TRANSITIONS_ENUM,
	}
	conf.Init(TEST_SENT)
	// [ROOT Economic news had little effect on financial markets .]
	//   0      1      2    3    4      5    6      7       8     9
	// Setup configuration:
	// C=(	[ROOT,had,effect], [.], A)
	// A={	(ROOT,	PRED,	had)
	// 		(had,	OBJ,	effect)
	// 		(effect,ATT,	little)
	//		(effect,ATT,	on)}
	predInd, _ := TEST_ENUM_RELATIONS.IndexOf(NLP.DepRel("PRED"))
	objInd, _ := TEST_ENUM_RELATIONS.IndexOf(NLP.DepRel("OBJ"))
	attInd, _ := TEST_ENUM_RELATIONS.IndexOf(NLP.DepRel("ATT"))
	conf.Nodes[0].AddModifier(3, predInd)
	conf.Nodes[3].Head = 0
	conf.Nodes[3].ELabel = predInd
	conf.Nodes[5].Head = 3
	conf.Nodes[4].Head = 5
	conf.Nodes[6].Head = 5
	conf.Nodes[3].AddModifier(5, objInd)
	conf.Nodes[5].ELabel = objInd
	conf.Nodes[5].AddModifier(4, attInd)
	conf.Nodes[4].ELabel = attInd
	conf.Nodes[5].AddModifier(6, attInd)
	conf.Nodes[6].ELabel = attInd

	// S=[ROOT,had,effect]
	// stack should be empty
	if _, sExists := conf.Stack().Peek(); sExists {
		panic("Initialized configuration should have empty stack")
	}
	conf.Stack().Push(3)
	conf.Stack().Push(5)

	// B=[.]
	conf.Queue().Clear()
	conf.Queue().Push(9)

	// A = {...}
	conf.Arcs().Add(&BasicDepArc{Head: 0, RawRelation: NLP.DepRel("PRED"), Modifier: 3})
	conf.Arcs().Add(&BasicDepArc{Head: 3, RawRelation: NLP.DepRel("OBJ"), Modifier: 5})
	conf.Arcs().Add(&BasicDepArc{Head: 5, RawRelation: NLP.DepRel("ATT"), Modifier: 4})
	conf.Arcs().Add(&BasicDepArc{Head: 5, RawRelation: NLP.DepRel("ATT"), Modifier: 6})

	return conf
}
