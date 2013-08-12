package Transition

import (
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP"
)

type ArcEager struct {
	ArcStandard
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
	switch transition[:2] {
	case "LA":
		wi, wiExists := conf.Stack().Pop()
		if wi == 0 {
			panic("Attempted to LA the root")
		}
		arcs := conf.Arcs().Get(&BasicDepArc{-1, "", wi})
		if len(arcs) > 0 {
			panic("Can't create arc for wi, it already has a head")
		}
		wj, wjExists := conf.Queue().Peek()
		if !(wiExists && wjExists) {
			panic("Can't LA, Stack and/or Queue are/is empty")
		}
		rel := DepRel(transition[3:])
		newArc := &BasicDepArc{wj, rel, wi}
		conf.Arcs().Add(newArc)
	case "RA":
		wi, wiExists := conf.Stack().Peek()
		wj, wjExists := conf.Queue().Pop()
		if !(wiExists && wjExists) {
			panic("Can't RA, Stack and/or Queue are/is empty")
		}
		rel := DepRel(transition[3:])
		newArc := &BasicDepArc{wi, rel, wj}
		conf.Stack().Push(wj)
		conf.Arcs().Add(newArc)
	case "RE":
		wi, wiExists := conf.Stack().Pop()
		arcs := conf.Arcs().Get(&BasicDepArc{-1, "", wi})
		if !wiExists {
			panic("Can't shift, queue is empty")
		}
		if len(arcs) == 0 {
			panic("Can't reduce wi if it doesn't have a head")
		}
	case "SH":
		wi, wiExists := conf.Queue().Pop()
		if !wiExists {
			panic("Can't shift, queue is empty")
		}
		conf.Stack().Push(wi)
	}
	conf.SetLastTransition(transition)
	return conf
}

func (a *ArcEager) TransitionTypes() []Transition {
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
	sPeek, sExists := conf.Stack().Peek()
	sPeekHasModifiers := len(conf.Arcs().Get(&BasicDepArc{-1, "", sPeek})) > 0
	if sExists {
		if sPeek != 0 && !sPeekHasModifiers {
			for _, rel := range a.Relations {
				transitions <- Transition("LA-" + rel)
			}
		}
	}
	if sExists && qExists {
		for _, rel := range a.Relations {
			transitions <- Transition("RA-" + rel)
		}
	}
	if sPeekHasModifiers {
		transitions <- Transition("RE")
	}
	if qExists {
		transitions <- Transition("SH")
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
		a.oracle = Oracle(&ArcEagerOracle{})
	}
}

type ArcEagerOracle struct {
	ArcStandardOracle
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
	if bExists && sExists {
		// test if should Left-Attach
		arcs := o.arcSet.Get(&BasicDepArc{bTop, "", sTop})
		if len(arcs) > 0 {
			arc := arcs[0]
			return Transition("LA-" + string(arc.GetRelation()))
		}

		// test if should Right-Attach
		arcs = o.arcSet.Get(&BasicDepArc{sTop, "", bTop})
		if len(arcs) > 0 {
			arc := arcs[0]
			return Transition("RA-" + string(arc.GetRelation()))
		}

		// test if should reduce

		// if modifier < sTop, REDUCE
		arcs = o.arcSet.Get(&BasicDepArc{bTop, "", -1})
		for _, arc := range arcs {
			if arc.GetModifier() < sTop {
				return Transition("RE")
			}
		}
		// if head < sTop, REDUCE
		arcs = o.arcSet.Get(&BasicDepArc{-1, "", bTop})
		for _, arc := range arcs {
			if arc.GetHead() < sTop {
				return Transition("RE")
			}
		}
	}
	return "SH"
}
