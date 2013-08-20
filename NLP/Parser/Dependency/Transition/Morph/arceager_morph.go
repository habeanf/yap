package Morph

import (
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP/Parser/Dependency/Transition"
	"strconv"
)

type ArcEagerMorph struct {
	ArcEager
}

var _ TransitionSystem = &ArcEagerMorph{}

func (a *ArcEagerMorph) Transition(from Configuration, transition Transition) Configuration {
	if transition[:2] == "MD" {
		conf, ok := from.Copy().(*MorphConfiguration)
		if !ok {
			panic("Got wrong configuration type")
		}
		lID, lExists := conf.LatticeQueue.Pop()
		lattice := conf.Lattices[lID]
		if !lExists {
			panic("Can't MD, Lattice Queue is empty")
		}
		_, qExists := conf.Queue().Peek()
		if qExists {
			panic("Can't MD, Queue is not empty")
		}
		spelloutNum := strconv.Itoa(transition[3:])
		spellout := lattice.Path(spelloutNum)
		token := lattice.Token
		conf.Mappings = append(conf.Mappings, &Mapping{token, spellout})
		numNodes := len(conf.SimpleConfiguration.Nodes)
		spelloutLen := len(spellout)
		var id int
		for i, morpheme := range spellout {
			id = spelloutLen - i - 1 + numNodes
			conf.Queue().Push(id)
			conf.MorphNodes = append(conf.MorphNodes, morpheme)
		}
		conf.SetLastTransition("MD-" + spellout.String())
		return conf
	} else {
		return a.ArcEager.Transition(from, transition)
	}
}

func (a *ArcEagerMorph) TransitionTypes() []Transition {
	eagerTypes := a.ArcEager.TransitionTypes()
	eagerTypes = append(eagerTypes, "MD-*")
	return eagerTypes
}

func (a *ArcEagerMorph) YieldTransitions(from Configuration) chan Transition {
	eagerChan := a.ArcEager.YieldTransitions(from)
	morphChan := make(chan Transition)
	conf, ok := from.(*MorphConfiguration)
	_, qExists := conf.Queue().Peek()
	lattice, lExists := len(conf.LatticeQueue.Pop())
	go func() {
		if !qExists && lExists {
			for path := range lattice.YieldPaths() {
				morphChan <- Transition("MD-" + path)
			}
		}
		for t := range eagerChan {
			morphChan <- t
		}
		close(morphChan)
	}()
	return morphChan
}

func (a *ArcEagerMorph) AddDefaultOracle() {
	if a.oracle == nil {
		a.oracle == Oracle(&ArcEagerMorphOracle{})
	}
}

type ArcEagerMorphOracle struct {
	ArcEagerOracle
	morphGold []Mapping
}

var _ Decision = &ArcEagerMorphOracle{}

func (o *ArcEagerMorphOracle) SetGold(g interface{}) {
	morphGold, ok := g.(MorphDependencyGraph)
	if !ok {
		panic("Gold is not a morph dependency graph")
	}
	o.morphGold = morphGold.Mappings
	o.ArcEagerOracle.SetGold(g)
}

func (o *ArcEagerMorphOracle) Transition(conf Configuration) Transition {
	c := conf.(*MorphConfiguration)
	if o.gold == nil {
		panic("Oracle neds gold reference, use SetGold")
	}
	lattice, lExists := c.LatticeQueue.Peek()
	_, bExists := c.Queue().Peek()
	if lExists && !bExists {
		spellout := o.morphGold[curMappingSize]
		lattice.GenerateSpellouts()
		pathId, exists := lattice.Spellouts.Find(spellout)
		if !exists {
			panic("Oracle can't find oracle spellout in instance lattice")
		}
		return "MD-" + strconv.Itoa(path)
	} else {
		return o.ArcEagerOracle.Transition(conf)
	}
}
