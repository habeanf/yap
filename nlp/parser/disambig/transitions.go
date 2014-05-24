package disambig

import (
	. "chukuparser/algorithm/transition"
	. "chukuparser/nlp/types"
	"chukuparser/util"
	// "fmt"
	// "log"
)

type MDParam func(s Spellout) string

type MDTrans struct {
	ParamFunc MDParam

	Transitions *util.EnumSet
	oracle      Oracle
}

var _ TransitionSystem = &MDTrans{}

func (t *MDTrans) Transition(from Configuration, transition Transition) Configuration {
	c := from.Copy().(*MDConfig)

	paramStr := t.Transitions.ValueOf(int(transition))
	qTop, qExists := c.LatticeQueue.Pop()
	if !qExists {
		panic("Lattice queue is empty! Whatcha doin'?!")
	}

	lattice := c.Lattices[qTop]
	for _, spellout := range lattice.Spellouts {
		if t.ParamFunc(spellout) == paramStr {
			c.Mappings = append(c.Mappings, &Mapping{lattice.Token, spellout})
			c.SetLastTransition(transition)
			return c
		}
	}
	panic("given spellout not in lattice :`(")
}

func (t *MDTrans) TransitionTypes() []string {
	return []string{"MD-*"}
}

func (t *MDTrans) possibleTransitions(from Configuration, transitions chan Transition) {
	var transition int

	conf, ok := from.(*MDConfig)
	if !ok {
		panic("Got wrong configuration type")
	}

	qTop, qExists := conf.LatticeQueue.Peek()
	if qExists {
		lat := conf.Lattices[qTop]
		for _, spellout := range lat.Spellouts {
			transition, _ = t.Transitions.Add(t.ParamFunc(spellout))
			transitions <- Transition(transition)
		}
	}
	close(transitions)
}

func (t *MDTrans) YieldTransitions(conf Configuration) chan Transition {
	transitions := make(chan Transition)
	go t.possibleTransitions(conf, transitions)
	return transitions
}

func (t *MDTrans) Oracle() Oracle {
	return t.oracle
}

func (t *MDTrans) AddDefaultOracle() {
	t.oracle = &MDOracle{
		Transitions: t.Transitions,
		ParamFunc:   t.ParamFunc,
	}
}

func (t *MDTrans) Name() string {
	return "Standalone Morphological Disambiguator"
}

type MDOracle struct {
	Transitions *util.EnumSet
	gold        Mappings
	ParamFunc   MDParam
}

var _ Decision = &MDOracle{}

func (o *MDOracle) SetGold(g interface{}) {
	mappings, ok := g.(Mappings)
	if !ok {
		panic("Gold is not an array of mappings")
	}
	o.gold = mappings
}

func (o *MDOracle) Transition(conf Configuration) Transition {
	c := conf.(*MDConfig)

	if o.gold == nil {
		panic("Oracle needs gold reference, use SetGold")
	}

	qTop, qExists := c.LatticeQueue.Peek()
	if !qExists {
		panic("No lattices in given configuration to disambiguate")
	}
	if len(o.gold) <= qTop {
		panic("Gold has less mappings than given configuration")
	}
	spellout := o.gold[qTop].Spellout
	paramVal := o.ParamFunc(spellout)
	transition, _ := o.Transitions.Add(paramVal)
	return Transition(transition)
}

func (o *MDOracle) Name() string {
	return "MD Exact Match"
}
