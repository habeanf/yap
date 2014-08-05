package joint

import (
	. "chukuparser/alg/transition"
	. "chukuparser/nlp/types"
	"chukuparser/util"

	dep "chukuparser/nlp/parser/dependency/transition"
	morph "chukuparser/nlp/parser/dependency/transition/morph"
	"chukuparser/nlp/parser/disambig"
	// "fmt"
	// "log"
)

var TSAllOut bool

type JointTrans struct {
	MDTrans       TransitionSystem
	ArcSys        TransitionSystem
	Transitions   *util.EnumSet
	oracle        Oracle
	JointStrategy string
	MDTransition  Transition
}

var _ TransitionSystem = &JointTrans{}

func (t *JointTrans) Transition(from Configuration, transition Transition) Configuration {
	// TODO: inefficient double copying of internal configurations by underlying
	// transition systems
	c := from.Copy().(*JointConfig)
	if transition >= t.MDTransition {
		c.MDConfig = *t.MDTrans.Transition(&c.MDConfig, transition).(*disambig.MDConfig)
		// enqueue last disambiguated morpheme
		// and add as "node"
		nodeId := len(c.Nodes)
		if nodeId != len(c.MDConfig.Morphemes)-1 {
			panic("Mismatch between Nodes and Morphemes")
		}
		curMorpheme := c.MDConfig.Morphemes[nodeId]
		c.SimpleConfiguration.Queue().Enqueue(nodeId)
		newNode := &dep.TaggedDepNode{
			nodeId,
			curMorpheme.EForm,
			curMorpheme.EPOS,
			curMorpheme.EFCPOS,
			curMorpheme.EMHost,
			curMorpheme.EMSuffix,
			curMorpheme.Form,
			curMorpheme.POS,
		}
		// TODO: ask reut about swallowed/aglutinated ("H") morpheme as a node in the graph?
		c.SimpleConfiguration.Nodes = append(c.SimpleConfiguration.Nodes,
			dep.NewArcCachedDepNode(DepNode(newNode)))

	} else {
		c.SimpleConfiguration = *t.ArcSys.Transition(&c.SimpleConfiguration, transition).(*dep.SimpleConfiguration)
	}
	// paramStr := t.Transitions.ValueOf(int(transition))
	return c
}

func (t *JointTrans) TransitionTypes() []string {
	return append(t.MDTrans.TransitionTypes(), t.ArcSys.TransitionTypes()...)
}

func (t *JointTrans) TransitionStrategy(c *JointConfig) (shouldMD bool, shouldDep bool) {
	shouldMD = false
	shouldDep = false
	switch t.JointStrategy {
	case "":
		shouldMD = true
		shouldDep = true
	case "All":
		shouldMD = true
		shouldDep = true
	case "MDFirst":
		if !c.MDConfig.Terminal() {
			shouldMD = true
		} else {
			shouldDep = true
		}
		// case "Min3":
		// case "FinishLattice":
	}
	if !(shouldMD || shouldDep) && !(c.MDConfig.Terminal() && c.SimpleConfiguration.Terminal()) {
		panic("One of the underlying configurations is not terminal but no transition type specified")
	}
	return
}

func (t *JointTrans) YieldTransitions(conf Configuration) chan Transition {
	transitions := make(chan Transition)
	go func() {
		c := conf.(*JointConfig)
		shouldMD, shouldDep := t.TransitionStrategy(c)
		if shouldMD {
			mdTransitions := t.MDTrans.YieldTransitions(&c.MDConfig)
			for t := range mdTransitions {
				transitions <- t
			}
		}
		if shouldDep {
			depTransitions := t.ArcSys.YieldTransitions(&c.SimpleConfiguration)
			for t := range depTransitions {
				transitions <- t
			}
		}
	}()
	return transitions
}

func (t *JointTrans) Oracle() Oracle {
	return t.oracle
}

func (t *JointTrans) AddDefaultOracle() {
	t.oracle = &JointOracle{
		JointStrategy: t.JointStrategy,
		MDOracle:      t.MDTrans.Oracle(),
		ArcSysOracle:  t.ArcSys.Oracle(),
	}
}

func (t *JointTrans) Name() string {
	return "Joint Morpho-Syntactic - Strategy:" + t.JointStrategy
}

type JointOracle struct {
	gold           Mappings
	MDOracle       Oracle
	ArcSysOracle   Oracle
	JointStrategy  string
	OracleStrategy string
}

var _ Decision = &JointOracle{}

func (o *JointOracle) SetGold(g interface{}) {
	graph, ok := g.(*morph.BasicMorphGraph)
	if !ok {
		panic("Gold is not a morph.BasicMorphGraph")
	}
	o.gold = graph
	o.MDOracle.SetGold(graph.Mappings)
	o.ArcSysOracle.SetGold(graph.BasicDepGraph)
}

func (o *JointOracle) MDFirst(conf Configuration) Transition {
	c, ok := conf.(*JointConfig)
	if !ok {
		panic("Conf must be *JointConfig")
	}
	if !c.MDConfig.Terminal() {
		return o.MDOracle.Transition(&c.MDConfig)
	} else {
		return o.ArcSysOracle.Transition(&c.SimpleConfiguration)
	}
}

func (o *JointOracle) ArcGreedy(conf Configuration) Transition {
	c, ok := conf.(*JointConfig)
	if !ok {
		panic("Conf must be *JointConfig")
	}
	if c.SimpleConfiguration.Queue().Size() >= 3 && !c.MDConfig.Terminal() {
		return o.ArcSysOracle.Transition(&c.SimpleConfiguration)
	} else {
		return o.MDOracle.Transition(&c.MDConfig)
	}
}

func (o *JointOracle) Transition(conf Configuration) Transition {
	if o.gold == nil {
		panic("Oracle needs gold reference, use SetGold")
	}

	switch o.OracleStrategy {
	case "MDFirst":
		return o.MDFirst(conf)
	case "ArcGreedy":
		return o.ArcGreedy(conf)
	}

	return 0
}

func (o *JointOracle) Name() string {
	return "Joint Morpho-Syntactic - Strategy:" + o.OracleStrategy
}
