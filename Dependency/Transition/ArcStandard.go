package Transition

type ArcStandard struct {
	Relations []string
	Gold      *Graph
	Oracle    *OracleClassifier
}

func (a *ArcStandard) Transition(from *Configuration, transition string) *Configuration {
	conf := from.Copy()
	// Transition System:
	// LA-r	(S|wi,	wj|B,	A) => (S   ,	wj|B,	A+{(wj,r,wi)})	if: != 0
	// RA-r	(S|wi, 	wj|B,	A) => (S   ,	wi|B, 	A+{(wi,r,wj)})
	// SH	(S   ,	wi|B, 	A) => (S|wi,	   B,	A)
	switch transition[:2] {
	case "LA":
		wi, _ := conf.Stack().Pop()
		if wi == 0 {
			panic("Attempted to LA the root")
		}
		wj, _ := conf.Queue().Peek()
		rel := transition[3:]
		newArc := &DepArc{wj, rel, wi}
		conf.Arcs().Push(newArc)
	case "RA":
		wi, _ := conf.Stack().Pop()
		wj, _ := conf.Queue().Pop()
		rel := transition[3:]
		newArc := &DepArc{wi, rel, wj}
		conf.Queue().Push(wi)
		conf.Arcs().Push(newArc)
	case "SH":
		wi := conf.Queue().Pop()
		conf.Stack().Push(wi)
	}
	conf.SetLastTransition(transition)
	return conf
}

func (a *ArcStandard) TransitionSet() []string {
	all := make([]string, 2*len(a.Relations)+1)
	for i, rel := range a.Relations {
		all[i] = "LA-" + rel
		all[i+len(a.Relations)] = "RA-" + rel
	}
	all[len(a.Relations)*2] = "SHIFT"
}

func (a *ArcStandard) TransitionTypes() []string {
	return [...]string{"LA-*", "RA-*", "SHIFT"}
}

func (a *ArcStandard) Projective() bool {
	return true
}

func (a *ArcStandard) Labeled() bool {
	return true
}

func (a *ArcStandard) Oracle() *Decision {
	if a.Gold == nil {
		panic("Oracle can't make a decision without Gold data")
	}
	if a.Oracle == nil {
		a.Oracle = &OracleFunction{a.Gold}
	}
	return a.Oracle
}

type OracleFunction struct {
	gold   *Graph
	arcSet *ArcSet
}

func (o *OracleFunction) SetGold(g *Gold) {
	o.gold = g
	o.arcSet = NewArcSet(g)
}

func (o *OracleFunction) GetTransition(c *Configuration) string {
	// Given Gd=(Vd,Ad) # gold dependencies
	// o(c = (S,B,A)) =
	// LA-r	if	(B[0],r,S[0]) in Ad
	// RA-r	if	(S[0],r,B[0]) in Ad; and for all w,r', if (B[0],r',w) in Ad then (B[0],r',w) in A
	// SH	otherwise
	bTop, bExists := c.Queue().Peek()
	sTop, sExists := c.Stack().Peek()
	if bExists && sExists {
		// test if should Left-Attach
		arcs, exists := o.arcSet.Get(&DepArc{bTop, "", sTop})
		if exists {
			return "LA-" + arcs[0].Relation
		}

		// test if should Right-Attach
		arcs, exists := o.arcSet.Get(&DepArc{sTop, "", bTop})
		if exists {
			reverseArcs, _ := o.arcSet.Get(&DepArc{bTop, "", -1})
			for _, arc := range reverseArcs {
				_, revExists := c.Arcs().Get(arc)
				if !revExists {
					return "SH"
				}
			}
			return "RA-" + arcs[0]
		}
	}
	return "SH"
}
