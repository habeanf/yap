package Transition

import (
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP"
)

type ArcStandard struct {
	Relations []string
	oracle    *Decision
}

// Verify that ArcStandard is a TransitionSystem
var _ TransitionSystem = &ArcStandard{}

func (a *ArcStandard) Transition(abstractFrom interface{}, transition Transition) *Configuration {
	from := abstractFrom.(*SimpleConfiguration)
	conf := (*from).Copy().(*SimpleConfiguration)
	// Transition System:
	// LA-r	(S|wi,	wj|B,	A) => (S   ,	wj|B,	A+{(wj,r,wi)})	if: i != 0
	// RA-r	(S|wi, 	wj|B,	A) => (S   ,	wi|B, 	A+{(wi,r,wj)})
	// SH	(S   ,	wi|B, 	A) => (S|wi,	   B,	A)
	switch transition[:2] {
	case "LA":
		wi, _ := conf.Stack.Pop()
		if wi == 0 {
			panic("Attempted to LA the root")
		}
		wj, _ := conf.Queue.Peek()
		rel := transition[3:]
		newArc := &DepArc{wj, rel, wi}
		conf.Arcs().Add(newArc)
	case "RA":
		wi, _ := conf.Stack.Pop()
		wj, _ := conf.Queue.Dequeue()
		rel := transition[3:]
		newArc := &DepArc{wi, rel, wj}
		conf.Queue.Push(wi)
		conf.Arcs.Add(newArc)
	case "SH":
		wi := conf.Queue.Dequeue()
		conf.Stack.Push(wi)
	}
	conf.SetLastTransition(transition)
	return conf
}

func (a *ArcStandard) TransitionTypes() []Transition {
	return [...]string{"LA-*", "RA-*", "SHIFT"}
}

func (a *ArcStandard) Projective() bool {
	return true
}

func (a *ArcStandard) Labeled() bool {
	return true
}

func (a *ArcStandard) Oracle() *Decision {
	return a.oracle
}

func (a *ArcStandard) AddDefaultOracle() {
	if a.Oracle == nil {
		a.Oracle = new(OracleFunction)
	}
}

type OracleFunction struct {
	gold   *DependencyGraph
	arcSet *ArcSet
}

func (o *OracleFunction) SetGold(g *DependencyGraph) {
	o.gold = g
	o.arcSet = NewArcSet(g)
}

func (o *OracleFunction) GetTransition(c *Configuration) string {
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
	if bExists && sExists {
		// test if should Left-Attach
		arcs := o.arcSet.Get(DepArc{bTop, "", sTop})
		if len(arcs) > 0 {
			return "LA-" + arcs[0].Relation
		}

		// test if should Right-Attach
		arcs = o.arcSet.Get(DepArc{sTop, "", bTop})
		if len(arc) > 0 {
			reverseArcs := o.arcSet.Get(DepArc{bTop, "", -1})
			// for all w,r', if (B[0],r',w) in Ad then (B[0],r',w) in A
			// otherwise, return SH
			for _, arc := range reverseArcs {
				revArcs := c.Arcs().Get(arc)
				if len(revArcs) == 0 {
					return "SH"
				}
			}
			return "RA-" + arcs[0].Relation
		}
	}
	return "SH"
}
