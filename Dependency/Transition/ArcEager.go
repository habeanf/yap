package Transition

type ArcEager struct {
	ArcStandard
}

func (a *ArcEager) Transition(from *Configuration, transition string) *Configuration {
	conf := from.Copy()
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
		arcs := conf.Arcs().Get(DepArc{-1, "", wi})
		if len(arcs) > 0 {
			panic("Can't create arc for wi, it already has a head (CHECK YO'SELF!)")
		}
		wj, _ := conf.Queue().Peek()
		rel := transition[3:]
		newArc := DepArc{wj, rel, wi}
		conf.Arcs().Add(newArc)
	case "RA":
		wi, _ := conf.Stack().Peek()
		wj, _ := conf.Queue().Dequeue()
		rel := transition[3:]
		newArc := DepArc{wi, rel, wj}
		conf.Stack().Push(wj)
		conf.Arcs().Add(newArc)
	case "RE":
		wi, _ := conf.Stack().Pop()
		arcs := conf.Arcs().Get(DepArc{-1, "", wi})
		if len(arcs) == 0 {
			panic("Can't reduce wi if it doesn't have a head")
		}
	case "SH":
		wi, _ := conf.Queue().Dequeue()
		conf.Stack().Push(wi)
	}
	conf.SetLastTransition(transition)
	return conf
}

func (a *ArcEager) TransitionSet() []string {
	standardSet := (a.(*ArcStandard)).TransitionSet()
	standardSet = append(standardSet, "RE")
	return standardSet
}

func (a *ArcEager) TransitionTypes() []string {
	standardTypes := (a.(*ArcStandard)).TransitionTypes()
	standardTypes = append(standardTypes, "RE")
	return standardTypes
}

func (a *ArcEager) Oracle() *Decision {
	if a.Gold == nil {
		panic("Oracle can't make a decision without Gold data")
	}
	panic("Oracle not implemented yet")
}
