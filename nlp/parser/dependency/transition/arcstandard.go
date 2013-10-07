package Transition

import (
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP/Types"
	"chukuparser/Util"
)

type ArcStandard struct {
	oracle             Oracle
	Relations          *Util.EnumSet
	Transitions        *Util.EnumSet
	SHIFT, LEFT, RIGHT Transition
}

// Verify that ArcStandard is a TransitionSystem
var _ TransitionSystem = &ArcStandard{}

func (a *ArcStandard) Transition(from Configuration, transition Transition) Configuration {
	conf, ok := from.Copy().(*SimpleConfiguration)
	if !ok {
		panic("Got wrong configuration type")
	}
	// Transition System:
	// LA-r	(S|wi,	wj|B,	A) => (S   ,	wj|B,	A+{(wj,r,wi)})	if: i != 0
	// RA-r	(S|wi, 	wj|B,	A) => (S   ,	wi|B, 	A+{(wi,r,wj)})
	// SH	(S   ,	wi|B, 	A) => (S|wi,	   B,	A)
	// switch transition[:2] {
	switch {
	// case "LA":
	case transition >= a.LEFT && transition < a.RIGHT:
		wi, wiExists := conf.Stack().Pop()
		if wi == 0 {
			panic("Attempted to LA the root")
		}
		wj, wjExists := conf.Queue().Peek()
		if !(wiExists && wjExists) {
			panic("Can't LA, Stack and/or Queue are/is empty")
		}
		// relation := DepRel(transition[3:])
		relation := int(transition - a.LEFT)
		relationValue := a.Relations.ValueOf(relation).(DepRel)
		newArc := &BasicDepArc{wj, relation, wi, relationValue}
		conf.Arcs().Add(newArc)
	// case "RA":
	case transition >= a.RIGHT:
		wi, wiExists := conf.Stack().Pop()
		wj, wjExists := conf.Queue().Pop()
		if !(wiExists && wjExists) {
			panic("Can't RA, Stack and/or Queue are/is empty")
		}
		// rel := DepRel(transition[3:])
		rel := int(transition - a.RIGHT)
		relValue := a.Relations.ValueOf(rel).(DepRel)
		newArc := &BasicDepArc{wi, rel, wj, relValue}
		conf.Queue().Push(wi)
		conf.Arcs().Add(newArc)
	case transition == a.SHIFT:
		wi, wiExists := conf.Queue().Pop()
		if !wiExists {
			panic("Can't shift, queue is empty")
		}
		conf.Stack().Push(wi)
	default:
	}
	conf.SetLastTransition(transition)
	return conf
}

func (a *ArcStandard) possibleTransitions(from Configuration, transitions chan Transition) {
	conf, ok := from.(*SimpleConfiguration)
	if !ok {
		panic("Got wrong configuration type")
	}
	_, qExists := conf.Queue().Peek()
	sPeek, sExists := conf.Stack().Peek()
	if qExists {
		transitions <- Transition(a.SHIFT)
	}
	if sExists {
		if sPeek != 0 {
			for rel, _ := range a.Relations.Index {
				// transitions <- Transition("LA-" + rel)
				transitions <- Transition(int(a.LEFT) + rel)
			}
		}
	}
	if sExists && qExists {
		for rel, _ := range a.Relations.Index {
			// transitions <- Transition("RA-" + rel)
			transitions <- Transition(int(a.RIGHT) + rel)
		}
	}
	close(transitions)
}

func (a *ArcStandard) YieldTransitions(from Configuration) chan Transition {
	transitions := make(chan Transition)
	go a.possibleTransitions(from, transitions)
	return transitions
}

func (a *ArcStandard) TransitionTypes() []string {
	return []string{"LA-*", "RA-*", "SH"}
}

func (a *ArcStandard) Projective() bool {
	return true
}

func (a *ArcStandard) Labeled() bool {
	return true
}

func (a *ArcStandard) Oracle() Oracle {
	return a.oracle
}

func (a *ArcStandard) AddDefaultOracle() {
	if a.oracle == nil {
		a.oracle = Oracle(&ArcStandardOracle{Transitions: a.Transitions})
	}
}

type ArcStandardOracle struct {
	Transitions *Util.EnumSet
	gold        LabeledDependencyGraph
	arcSet      *ArcSetSimple
}

var _ Decision = &ArcStandardOracle{}

func (o *ArcStandardOracle) SetGold(g interface{}) {
	labeledGold, ok := g.(LabeledDependencyGraph)
	if !ok {
		panic("Gold is not a labeled dependency graph")
	}
	o.gold = labeledGold
	o.arcSet = NewArcSetSimpleFromGraph(o.gold)
}

func (o *ArcStandardOracle) Transition(conf Configuration) Transition {
	c := conf.(*SimpleConfiguration)

	if o.gold == nil {
		panic("Oracle needs gold reference, use SetGold")
	}
	// Given Gd=(Vd,Ad) # gold dependencies
	// o(c = (S,B,A)) =
	// LA-r	if	(B[0],r,S[0]) in Ad
	// RA-r	if	(S[0],r,B[0]) in Ad; and for all w,r', if (B[0],r',w) in Ad then (B[0],r',w) in A
	// SH	otherwise
	bTop, bExists := c.Queue().Peek()
	sTop, sExists := c.Stack().Peek()
	var index int
	if bExists && sExists {
		// test if should Left-Attach
		arcs := o.arcSet.Get(&BasicDepArc{bTop, -1, sTop, DepRel("")})
		if len(arcs) > 0 {
			arc := arcs[0]
			index, _ = o.Transitions.IndexOf(DepRel("LA-" + string(arc.GetRelation())))
			return Transition(index)
		}

		// test if should Right-Attach
		arcs = o.arcSet.Get(&BasicDepArc{sTop, -1, bTop, DepRel("")})
		if len(arcs) > 0 {
			reverseArcs := o.arcSet.Get(&BasicDepArc{bTop, -1, -1, DepRel("")})
			// for all w,r', if (B[0],r',w) in Ad then (B[0],r',w) in A
			// otherwise, return SH
			for _, arc := range reverseArcs {
				revArcs := c.Arcs().Get(arc)
				if len(revArcs) == 0 {
					index, _ = o.Transitions.IndexOf(DepRel("SH"))
					return Transition(index)
				}
			}
			arc := arcs[0]
			// return Transition("RA-" + string(arc.GetRelation()))
			index, _ = o.Transitions.IndexOf(DepRel("RA-" + string(arc.GetRelation())))
			return Transition(index)
		}
	}
	index, _ = o.Transitions.IndexOf(DepRel("SH"))
	return Transition(index)
}
