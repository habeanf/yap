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
			panic("Attempted to LA the root (Y U NO CHECK PRECONDITION?!)")
		}
		arcs := conf.Arcs().Get(&BasicDepArc{wi, "", -1})
		if len(arcs) > 0 {
			panic("Can't create arc for wi, it already has a head (CHECK YO'SELF!)")
		}
		wj, _ := conf.Queue().Peek()
		rel := DepRel(transition[3:])
		newArc := &BasicDepArc{wi, rel, wj}
		conf.Arcs().Add(newArc)
	case "RA":
		wi, _ := conf.Stack().Peek()
		wj, _ := conf.Queue().Pop()
		rel := DepRel(transition[3:])
		newArc := &BasicDepArc{wj, rel, wi}
		conf.Stack().Push(wj)
		conf.Arcs().Add(newArc)
	case "RE":
		wi, _ := conf.Stack().Pop()
		arcs := conf.Arcs().Get(&BasicDepArc{wi, "", -1})
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

func (a *ArcEager) Oracle() Oracle {
	return nil
}
