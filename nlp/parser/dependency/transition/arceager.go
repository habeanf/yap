package transition

import (
	. "chukuparser/algorithm/transition"
	. "chukuparser/nlp/types"
	"chukuparser/util"
	"fmt"
	// "log"
)

type ArcEager struct {
	ArcStandard
	POPROOT, REDUCE Transition
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
		// arcs := conf.Arcs().Get(&BasicDepArc{-1, -1, wi, DepRel("")})
		if conf.Arcs().HasHead(wi) {
			panic("Can't create arc for wi, it already has a head")
		}
		wj, wjExists := conf.Queue().Peek()
		if !(wiExists && wjExists) {
			panic("Can't LA, Stack and/or Queue are/is empty")
		}
		// relation := DepRel(transition[3:])
		relation := int(transition - a.LEFT)
		relationValue := a.Relations.ValueOf(relation).(DepRel)
		newArc := &BasicDepArc{wj, relation, wi, relationValue}
		conf.AddArc(newArc)
	// case "RA":
	case transition >= a.RIGHT:
		wi, wiExists := conf.Stack().Peek()
		wj, wjExists := conf.Queue().Pop()
		if !(wiExists && wjExists) {
			panic("Can't RA, Stack and/or Queue are/is empty")
		}
		// rel := DepRel(transition[3:])
		rel := int(transition - a.RIGHT)
		relValue := a.Relations.ValueOf(rel).(DepRel)
		newArc := &BasicDepArc{wi, rel, wj, relValue}
		conf.Stack().Push(wj)
		conf.AddArc(newArc)
	case transition == a.REDUCE:
		if conf.Stack().Size() == 1 {
			panic("Attempted to reduce to ROOT (should POPROOT)")
		}
		wi, wiExists := conf.Stack().Pop()
		// arcs := conf.Arcs().Get(&BasicDepArc{-1, -1, wi, DepRel("")})
		if !wiExists {
			panic("Can't reduce, queue is empty")
		}
		_, wjExists := conf.Queue().Peek()
		// if len(arcs) == 0 {
		if !conf.Arcs().HasHead(wi) && wjExists {
			panic(fmt.Sprintf("Can't reduce %d if it doesn't have a head", wi))
		}
	case transition == a.SHIFT:
		wi, wiExists := conf.Queue().Pop()
		if !wiExists {
			panic("Can't shift, queue is empty")
		}
		conf.Stack().Push(wi)
	case transition == a.POPROOT:
		_, wjExists := conf.Queue().Pop()
		if wjExists {
			panic("Can't poproot, queue is not empty")
		}
		stackSize := conf.Stack().Size()
		if stackSize != 1 {
			panic("Can't poproot, stack has doesn't have just 1 value")
		}
		wi, _ := conf.Stack().Pop()
		relID, _ := a.Relations.IndexOf(ROOT_LABEL)
		newArc := &BasicDepArc{0, relID, wi, DepRel(ROOT_LABEL)}
		conf.AddArc(newArc)
	}
	conf.SetLastTransition(transition)
	return conf
}

func (a *ArcEager) TransitionTypes() []string {
	standardTypes := a.ArcStandard.TransitionTypes()
	standardTypes = append(standardTypes, "RE")
	standardTypes = append(standardTypes, "PR")
	return standardTypes
}

func (a *ArcEager) possibleTransitions(from Configuration, transitions chan Transition) {
	conf, ok := from.(*SimpleConfiguration)
	if !ok {
		panic("Got wrong configuration type")
	}
	_, qExists := conf.Queue().Peek()
	sSize := conf.Stack().Size()
	if qExists {
		if conf.GetLastTransition() != a.REDUCE {
			transitions <- Transition(a.SHIFT)
		}
	} else {
		if sSize == 1 {
			transitions <- Transition(a.POPROOT)
		}
	}
	sPeek, sExists := conf.Stack().Peek()

	// sPeekHasModifiers2 := len(conf.Arcs().Get(&BasicDepArc{-1, -1, sPeek, DepRel("")})) > 0
	if sExists {
		if qExists {
			for rel, _ := range a.Relations.Index {
				// transitions <- Transition("RA-" + rel)
				transitions <- Transition(int(a.RIGHT) + rel)
			}
		}
		sPeekHasModifiers := conf.Arcs().HasHead(sPeek)
		if (sPeekHasModifiers || !qExists) && sSize > 1 {
			transitions <- Transition(a.REDUCE)
		}
		if qExists && !sPeekHasModifiers {
			for rel, _ := range a.Relations.Index {
				// transitions <- Transition("LA-" + rel)
				transitions <- Transition(int(a.LEFT) + rel)
			}
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
	Transitions *util.EnumSet
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
	var (
		index  int
		exists bool
		arcs   []LabeledDepArc
	)

	if sExists && !bExists {
		sSize := c.Stack().Size()
		if sSize == 1 {
			index, exists = o.Transitions.IndexOf("PR")
			if !exists {
				panic("PR not found in trans enum")
			}
			return Transition(index)
		}
	}

	if bExists && sExists {
		// test if should Left-Attach
		arcs = o.arcSet.Get(&BasicDepArc{bTop, -1, sTop, DepRel("")})
		// log.Println("LA test returned", len(arcs))
		if len(arcs) > 0 {
			arc := arcs[0]
			index, exists = o.Transitions.IndexOf("LA-" + string(arc.GetRelation()))
			if !exists {
				panic("LA-" + string(arc.GetRelation()) + " not found in trans enum")
			}
			// log.Println("Oracle", o.Transitions.ValueOf(index))
			return Transition(index)
		}

		// test if should Right-Attach
		arcs = o.arcSet.Get(&BasicDepArc{sTop, -1, bTop, DepRel("")})
		// log.Println("RA test returned", len(arcs))
		if len(arcs) > 0 {
			arc := arcs[0]
			index, exists = o.Transitions.IndexOf("RA-" + string(arc.GetRelation()))
			if !exists {
				panic("RA-" + string(arc.GetRelation()) + " not found in trans enum")
			}
			// log.Println("Oracle", o.Transitions.ValueOf(index))
			return Transition(index)
		}
	}
	// test if should reduce
	if sExists {
		// if modifier < sTop, REDUCE
		arcs = o.arcSet.Get(&BasicDepArc{bTop, -1, -1, DepRel("")})
		for _, arc := range arcs {
			if arc.GetModifier() < sTop {
				index, exists = o.Transitions.IndexOf("RE")
				if !exists {
					panic("RE not found in trans enum")
				}
				// log.Println("Oracle", o.Transitions.ValueOf(index), "arc", arc.String())
				return Transition(index)
			}
		}
		// if head < sTop, REDUCE
		arcs = o.arcSet.Get(&BasicDepArc{-1, -1, bTop, DepRel("")})
		for _, arc := range arcs {
			if arc.GetHead() < sTop {
				index, exists = o.Transitions.IndexOf("RE")
				if !exists {
					panic("RE not found in trans enum")
				}
				// log.Println("Oracle2", o.Transitions.ValueOf(index), "arc", arc.String())
				return Transition(index)
			}
		}
	}
	if bExists {
		index, exists = o.Transitions.IndexOf("SH")
		if !exists {
			panic("SH not found in trans enum")
		}
		return Transition(index)
	}
	panic(fmt.Sprintf("Oracle cannot take any action when both stack and queue are empty (%v,%v)", sExists, bExists))
}
