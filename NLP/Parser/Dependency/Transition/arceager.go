package Transition

import (
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP/Types"
	"chukuparser/Util"
	// "log"
)

type ArcEager struct {
	ArcStandard
	REDUCE Transition
}

// Verify that ArcEager is a TransitionSystem
var _ TransitionSystem = &ArcEager{}

func (a *ArcEager) Transition(from Configuration, transition Transition) Configuration {
	conf, ok := from.Copy().(*SimpleConfiguration)
	if !ok {
		panic("Got wrong configuration type")
	}
	// Transition System:
	// LA-r	(S|wi,	wj|B,	A) => (S      ,	wj|B,	A+{(wj,r,wi)})	if: (wk,r',wi) notin A; i != 0
	// RA-r	(S|wi,	wj|B,	A) => (S|wi|wj,	   B,	A+{(wi,r,wj)})
	// RE	(S|wi,	   B,	A) => (S      ,	   B,	A)				if: (wk,r',wi) in A
	// SH	(S   ,	wi|B, 	A) => (S|wi   ,	   B,	A)
	// switch transition[:2] {
	switch {
	case transition >= a.LEFT && transition < a.RIGHT:
		wi, wiExists := conf.Stack().Pop()
		if wi == 0 {
			panic("Attempted to LA the root")
		}
		arcs := conf.Arcs().Get(&BasicDepArc{-1, -1, wi, ""})
		if len(arcs) > 0 {
			panic("Can't create arc for wi, it already has a head")
		}
		wj, wjExists := conf.Queue().Peek()
		if !(wiExists && wjExists) {
			panic("Can't LA, Stack and/or Queue are/is empty")
		}
		// relation := DepRel(transition[3:])
		relation := int(transition - a.LEFT)
		relationValue := DepRel(a.Relations.ValueOf(relation).(string))
		newArc := &BasicDepArc{wj, relation, wi, relationValue}
		conf.Arcs().Add(newArc)
	// case "RA":
	case transition >= a.RIGHT:
		wi, wiExists := conf.Stack().Peek()
		wj, wjExists := conf.Queue().Pop()
		if !(wiExists && wjExists) {
			panic("Can't RA, Stack and/or Queue are/is empty")
		}
		// rel := DepRel(transition[3:])
		rel := int(transition - a.RIGHT)
		relValue := DepRel(a.Relations.ValueOf(rel).(string))
		newArc := &BasicDepArc{wi, rel, wj, relValue}
		conf.Stack().Push(wj)
		conf.Arcs().Add(newArc)
	case transition == a.REDUCE:
		wi, wiExists := conf.Stack().Pop()
		arcs := conf.Arcs().Get(&BasicDepArc{-1, -1, wi, ""})
		if !wiExists {
			panic("Can't shift, queue is empty")
		}
		if len(arcs) == 0 {
			panic("Can't reduce wi if it doesn't have a head")
		}
	case transition == a.SHIFT:
		wi, wiExists := conf.Queue().Pop()
		if !wiExists {
			panic("Can't shift, queue is empty")
		}
		conf.Stack().Push(wi)
	}
	conf.SetLastTransition(transition)
	return conf
}

func (a *ArcEager) TransitionTypes() []string {
	standardTypes := a.ArcStandard.TransitionTypes()
	standardTypes = append(standardTypes, "RE")
	return standardTypes
}

func (a *ArcEager) possibleTransitions(from Configuration, transitions chan Transition) {
	conf, ok := from.(*SimpleConfiguration)
	if !ok {
		panic("Got wrong configuration type")
	}
	_, qExists := conf.Queue().Peek()
	if qExists {
		transitions <- Transition(a.SHIFT)
	}
	sPeek, sExists := conf.Stack().Peek()
	if sExists && qExists {
		for rel, _ := range a.Relations.Index {
			// transitions <- Transition("RA-" + rel)
			transitions <- Transition(int(a.RIGHT) + rel)
		}
	}

	sPeekHasModifiers := len(conf.Arcs().Get(&BasicDepArc{-1, -1, sPeek, ""})) > 0
	if sPeekHasModifiers {
		transitions <- Transition(a.REDUCE)
	}
	if sExists && qExists && sPeek != 0 && !sPeekHasModifiers {
		for rel, _ := range a.Relations.Index {
			// transitions <- Transition("LA-" + rel)
			transitions <- Transition(int(a.LEFT) + rel)
		}
	}
	close(transitions)
}

func (a *ArcEager) YieldTransitions(from Configuration) chan Transition {
	transitions := make(chan Transition)
	go a.possibleTransitions(from, transitions)
	return transitions
}

func (a *ArcEager) AddDefaultOracle() {
	if a.oracle == nil {
		a.oracle = Oracle(&ArcEagerOracle{Transitions: a.Transitions})
	}
}

type ArcEagerOracle struct {
	ArcStandardOracle
	Transitions *Util.EnumSet
}

var _ Decision = &ArcEagerOracle{}

func (o *ArcEagerOracle) Transition(conf Configuration) Transition {
	c := conf.(*SimpleConfiguration)

	if o.gold == nil {
		panic("Oracle needs gold reference, use SetGold")
	}
	// # http://www.cs.bgu.ac.il/~yoavg/publications/coling2012dynamic.pdf
	// Given Gd=(Vd,Ad) # gold dependencies
	// o(c = (S,B,A)) =
	// LA-r	if	(B[0],r,S[0]) in Ad
	// RA-r	if	(S[0],r,B[0]) in Ad
	// RE	if	i=S[0],  j=B[0], exists k<i and exists r: (B[0],r,k) in Ad or (k,r,B[0]) in Ad
	// SH	otherwise
	bTop, bExists := c.Queue().Peek()
	sTop, sExists := c.Stack().Peek()
	var index int
	if bExists && sExists {
		// test if should Left-Attach
		arcs := o.arcSet.Get(&BasicDepArc{bTop, -1, sTop, ""})
		if len(arcs) > 0 {
			arc := arcs[0]
			index, _ = o.Transitions.IndexOf("LA-" + string(arc.GetRelation()))
			return Transition(index)
		}

		// test if should Right-Attach
		arcs = o.arcSet.Get(&BasicDepArc{sTop, -1, bTop, ""})
		if len(arcs) > 0 {
			arc := arcs[0]
			index, _ = o.Transitions.IndexOf("RA-" + string(arc.GetRelation()))
			return Transition(index)
		}

		// test if should reduce

		// if modifier < sTop, REDUCE
		arcs = o.arcSet.Get(&BasicDepArc{bTop, -1, -1, ""})
		for _, arc := range arcs {
			if arc.GetModifier() < sTop {
				index, _ = o.Transitions.IndexOf("RE")
				return Transition(index)
			}
		}
		// if head < sTop, REDUCE
		arcs = o.arcSet.Get(&BasicDepArc{-1, -1, bTop, ""})
		for _, arc := range arcs {
			if arc.GetHead() < sTop {
				index, _ = o.Transitions.IndexOf("RE")
				return Transition(index)
			}
		}
	}
	index, _ = o.Transitions.IndexOf("SH")
	return Transition(index)
}
