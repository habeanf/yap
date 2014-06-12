package disambig

import (
	. "chukuparser/algorithm/transition"
	. "chukuparser/nlp/types"
	"chukuparser/util"
	// "fmt"
	"log"
)

var TSAllOut bool

type MDTrans struct {
	ParamFunc MDParam

	Transitions *util.EnumSet
	oracle      Oracle
}

var _ TransitionSystem = &MDTrans{}

func (t *MDTrans) Transition(from Configuration, transition Transition) Configuration {
	c := from.Copy().(*MDConfig)

	paramStr := t.Transitions.ValueOf(int(transition))
	qTop, qExists := c.LatticeQueue.Peek()
	if !qExists {
		panic("Lattice queue is empty! Whatcha doin'?!")
	}

	if TSAllOut {
		log.Println("Qtop:", qTop, "currentNode", c.CurrentLatNode)
	}
	lattice := c.Lattices[qTop]
	// log.Println("At lattice", qTop, "-", lattice.Token)
	// log.Println("Current lat node", c.CurrentLatNode)
	nexts, _ := lattice.Next[c.CurrentLatNode]
	// log.Println("Nexts are", nexts)
	// log.Println("Morphemes are", lattice.Morphemes)
	for _, next := range nexts {
		morph := lattice.Morphemes[next]
		// log.Println("Comparing morpheme param val", t.ParamFunc(morph), "to", paramStr)
		if t.ParamFunc(morph) == paramStr {
			c.AddMapping(morph)
			c.SetLastTransition(transition)
			return c
		}
	}
	panic("transition did not match a given morpheme :`(")
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
		if conf.CurrentLatNode < lat.Top() {
			nextList, _ := lat.Next[conf.CurrentLatNode]
			for _, next := range nextList {
				transition, _ = t.Transitions.Add(t.ParamFunc(lat.Morphemes[next]))
				transitions <- Transition(transition)
			}
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
	goldSpellout := o.gold[qTop].Spellout

	confSpellout := c.Mappings[len(c.Mappings)-1].Spellout
	// log.Println("At lattice", qTop, "mapping", len(confSpellout))
	// currentMorph := goldSpellout[len(confSpellout)]
	// log.Println("Gold morpheme", currentMorph.Form)
	paramVal := o.ParamFunc(goldSpellout[len(confSpellout)])
	// log.Println("Gold transition", paramVal)
	transition, _ := o.Transitions.Add(paramVal)
	return Transition(transition)
}

func (o *MDOracle) Name() string {
	return "MD Exact Match"
}
