package disambig

import (
	. "yap/alg/transition"
	. "yap/nlp/types"
	"yap/util"

	"fmt"
	"log"
)

var TSAllOut bool

type MDTrans struct {
	ParamFunc MDParam
	POP       Transition

	Transitions *util.EnumSet
	oracle      Oracle

	Log    bool
	UsePOP bool
}

var _ TransitionSystem = &MDTrans{}

func (t *MDTrans) Transition(from Configuration, transition Transition) Configuration {
	c := from.Copy().(*MDConfig)

	if t.UsePOP && transition == t.POP {
		c.Pop()
		c.SetLastTransition(transition)
		if TSAllOut || t.Log {
			log.Println("POPing")
		}
		return c
	}
	if transition == Transition(0) {
		c.SetLastTransition(transition)
		if TSAllOut || t.Log {
			log.Println("Idling")
		}
		return c
	}
	paramStr := t.Transitions.ValueOf(int(transition))
	qTop, _ := c.LatticeQueue.Peek()
	// if !qExists {
	// 	panic("Lattice queue is empty! Whatcha doin'?!")
	// }

	if TSAllOut || t.Log {
		log.Println("Qtop:", qTop, "currentNode", c.CurrentLatNode)
	}
	lattice := c.Lattices[qTop]
	if TSAllOut || t.Log {
		log.Println("At lattice", qTop, "-", lattice.Token)
		log.Println("Current lat node", c.CurrentLatNode)
	}
	nexts, _ := lattice.Next[c.CurrentLatNode]
	if TSAllOut || t.Log {
		log.Println("Nexts are", nexts)
		log.Println("Morphemes are", lattice.Morphemes)
	}
	for _, next := range nexts {
		morph := lattice.Morphemes[next]
		if TSAllOut || t.Log {
			log.Println("Comparing morpheme param val", t.ParamFunc(morph), "to", paramStr)
		}
		if t.ParamFunc(morph) == paramStr {
			if TSAllOut || t.Log {
				log.Println("Adding morph", morph)
			}
			c.AddMapping(morph)
			c.SetLastTransition(transition)
			return c
		}
	}
	var panicStr string
	panicStr = "transition did not match a given morpheme :`( -- "
	panicStr += fmt.Sprintf("failed to transition to %v", paramStr)
	panic(panicStr)
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
	if t.UsePOP && ((!qExists && len(conf.Mappings) != conf.popped) ||
		(qExists && qTop != conf.popped)) {
		transitions <- t.POP
	} else {
		if qExists {
			lat := conf.Lattices[qTop]
			if conf.CurrentLatNode < lat.Top() {
				nextList, _ := lat.Next[conf.CurrentLatNode]
				if t.Log {
					log.Println("\t\tpossible transitions", nextList)
				}
				for _, next := range nextList {
					transition, _ = t.Transitions.Add(t.ParamFunc(lat.Morphemes[next]))
					transitions <- Transition(transition)
				}
			}
		} else {
			// if t.Log {
			// 	log.Println("\t\tpossible transitions IDLE")
			// }
			// transitions <- Transition(0)
		}
	}
	close(transitions)
}

func (a *MDTrans) GetTransitions(from Configuration) []int {
	retval := make([]int, 0, 10)
	transitions := a.YieldTransitions(from)
	for transition := range transitions {
		retval = append(retval, int(transition))
	}
	return retval
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
		UsePOP:      t.UsePOP,
	}
}

func (t *MDTrans) Name() string {
	return "Morpheme-Based Morphological Disambiguator"
}

type MDOracle struct {
	Transitions *util.EnumSet
	gold        Mappings
	ParamFunc   MDParam
	UsePOP      bool
}

var _ Decision = &MDOracle{}

func (o *MDOracle) SetGold(g interface{}) {
	mappings, ok := g.(Mappings)
	if !ok {
		panic("Gold is not an array of mappings")
	}
	o.gold = mappings
}

func (o *MDOracle) CountMatchingTrans(c *MDConfig, pf MDParam, testTrans string) (matches int, matching string) {
	qTop, _ := c.LatticeQueue.Peek()
	lat := c.Lattices[qTop]
	if c.CurrentLatNode >= lat.Top() {
		panic("current lat node >= lattice's top :s")
	}
	nextList, _ := lat.Next[c.CurrentLatNode]
	// log.Println("\t\tpossible transitions", nextList, "for", testTrans)
	for _, next := range nextList {
		transStr := pf(lat.Morphemes[next])
		if transStr == testTrans {
			// log.Println("\t\t\t", transStr, "matches")
			matching = o.ParamFunc(lat.Morphemes[next])
			matches++
			continue
		}
		// log.Println("\t\t\t", transStr)
	}
	return
}

func (o *MDOracle) Transition(conf Configuration) Transition {
	c := conf.(*MDConfig)

	if o.gold == nil {
		panic("Oracle needs gold reference, use SetGold")
	}

	qTop, qExists := c.LatticeQueue.Peek()
	if o.UsePOP && ((!qExists && len(c.Mappings) != c.popped) ||
		(qExists && qTop != c.popped)) {
		return c.POP
	}
	if !qExists {
		// oracle forces a single final idle
		// if c.Last != Transition(0) {
		// 	return Transition(0)
		// }
		panic("No lattices in given configuration to disambiguate")
	}
	if len(o.gold) <= qTop {
		panic("Gold has less mappings than given configuration")
	}
	goldSpellout := o.gold[qTop].Spellout
	// log.Println("Current gold")
	// log.Println(o.gold)
	// log.Println("Top", qTop)
	// log.Println("Yielding", goldSpellout)

	// log.Println("Current mappings")
	// log.Println(c.Mappings)
	var spellOutMorph int
	if len(c.Mappings) > 0 {
		confSpellout := c.Mappings[len(c.Mappings)-1].Spellout
		spellOutMorph = len(confSpellout)
		// log.Println("Confspellout")
		// log.Println(confSpellout)
		// log.Println("At lattice", qTop, "mapping", len(confSpellout))
		// log.Println("GoldSpellout", goldSpellout)
		// log.Println("len(confSpellout)", len(confSpellout))
		// currentMorph := goldSpellout[len(confSpellout)]
		// log.Println("Gold morpheme", currentMorph.Form)
	}

	var paramVal string
	if spellOutMorph < len(goldSpellout) {
		paramVal = o.ParamFunc(goldSpellout[spellOutMorph])
	} else {
		qTop, _ := c.LatticeQueue.Peek()
		lat := c.Lattices[qTop]
		nextList, _ := lat.Next[c.CurrentLatNode]
		if len(nextList) == 1 {
			log.Println("\t\tOracle has no gold, using only possible morpheme")
		} else {
			log.Println("\t\tOracle has no gold, arbitrarily attempting to use first possible morpheme")
		}
		paramVal = o.ParamFunc(lat.Morphemes[nextList[0]])
	}

	failoverPFs := []MDParam{Funcs_Main_POS, POS_Prop, POS, Form}
	verifyPossibleTransition := true
	if verifyPossibleTransition {
		matches, matching := o.CountMatchingTrans(c, o.ParamFunc, paramVal)
		if matches == 0 {
			// log.Println("\tmatch not found, trying to match relaxed param func")
			for _, relaxedPF := range failoverPFs {
				paramVal = relaxedPF(goldSpellout[spellOutMorph])
				matches, matching = o.CountMatchingTrans(c, relaxedPF, paramVal)
				if matches >= 1 {
					paramVal = matching
					break
				}
			}
		}
		if matches > 1 {
			log.Println("\t\tOracle found too many matches, arbitrarily designating last found match for token", qTop, ":", matching)
			// panic("found too many matches, can't distinguish gold morpheme")
		}
		paramVal = matching
		if matches == 0 {
			qTop, _ := c.LatticeQueue.Peek()
			lat := c.Lattices[qTop]
			nextList, _ := lat.Next[c.CurrentLatNode]
			if len(nextList) == 1 && lat.Morphemes[nextList[0]].CPOS == "NNP" {
				log.Println("\t\tOracle found no matches, only morpheme is NNP, assuming OOV for token", qTop, ":", matching)
				paramVal = o.ParamFunc(lat.Morphemes[nextList[0]])
				// log.Println("\t\tUsing transition", paramVal)
			} else {
				panic(fmt.Sprintf("failed to find gold match for gold morph %v", goldSpellout[spellOutMorph]))
			}
		} else {
			// log.Println("\tmatch found")
		}
	}

	// log.Println("Gold transition", paramVal)
	transition, _ := o.Transitions.Add(paramVal)
	return Transition(transition)
}

func (o *MDOracle) Name() string {
	return "MD Exact Match"
}
