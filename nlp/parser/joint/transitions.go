package joint

import (
	. "chukuparser/alg/transition"
	. "chukuparser/nlp/types"
	"chukuparser/util"
	// "fmt"
	// "log"
)

var TSAllOut bool

type JointTrans struct {
	MDTrans     TransitionSystem
	ArcSys      TransitionSystem
	Transitions *util.EnumSet
	oracle      Oracle
}

var _ TransitionSystem = &JointTrans{}

func (t *JointTrans) Transition(from Configuration, transition Transition) Configuration {
	c := from.Copy().(*JointConfig)

	// paramStr := t.Transitions.ValueOf(int(transition))
	return c
}

func (t *JointTrans) TransitionTypes() []string {
	return append(t.MDTrans.TransitionTypes(), t.ArcSys.TransitionTypes()...)
}

func (t *JointTrans) YieldTransitions(conf Configuration) chan Transition {
	transitions := make(chan Transition)
	// go t.possibleTransitions(conf, transitions)
	return transitions
}

func (t *JointTrans) Oracle() Oracle {
	return t.oracle
}

func (t *JointTrans) AddDefaultOracle() {
	t.oracle = &JointOracle{}
}

func (t *JointTrans) Name() string {
	return "Joint Morpho-Syntactic Transition System"
}

type JointOracle struct {
	gold Mappings
}

var _ Decision = &JointOracle{}

func (o *JointOracle) SetGold(g interface{}) {
	mappings, ok := g.(Mappings)
	if !ok {
		panic("Gold is not an array of mappings")
	}
	o.gold = mappings
}

func (o *JointOracle) Transition(conf Configuration) Transition {
	// c := conf.(*JointConfig)

	// if o.gold == nil {
	// 	panic("Oracle needs gold reference, use SetGold")
	// }
	return 0

}

func (o *JointOracle) Name() string {
	return "Joint Morpho-Syntactic Oracle"
}
