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
		wi, _ := conf.Stack().Pop()
		if wi == 0 {
			panic("Attempted to LA the root")
		}
		arcs := conf.Arcs().Get(&BasicDepArc{-1, "", wi})
		if len(arcs) > 0 {
			panic("Can't create arc for wi, it already has a head")
		}
		wj, _ := conf.Queue().Peek()
		rel := DepRel(transition[3:])
		newArc := &BasicDepArc{wj, rel, wi}
		conf.Arcs().Add(newArc)
	case "RA":
		wi, _ := conf.Stack().Peek()
		wj, _ := conf.Queue().Pop()
		rel := DepRel(transition[3:])
		newArc := &BasicDepArc{wi, rel, wj}
		conf.Stack().Push(wj)
		conf.Arcs().Add(newArc)
	case "RE":
		wi, _ := conf.Stack().Pop()
		arcs := conf.Arcs().Get(&BasicDepArc{-1, "", wi})
		if len(arcs) == 0 {
			panic("Can't reduce wi if it doesn't have a head")
		}
	case "SH":
		wi, _ := conf.Queue().Pop()
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

func (a *ArcEager) AddDefaultOracle() {
	if a.oracle == nil {
		a.oracle = Oracle(&ArcEagerOracle{})
	}
}

type ArcEagerOracle struct {
	ArcStandardOracle
}

var _ Decision = &ArcEagerOracle{}

func (o *ArcEagerOracle) GetTransition(conf Configuration) Transition {
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
