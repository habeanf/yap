package Morph

import (
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
	"strconv"

	// "log"
)

type ArcEagerMorph struct {
	ArcEager
	oracle Oracle
}

var _ TransitionSystem = &ArcEagerMorph{}

func (a *ArcEagerMorph) Transition(from Configuration, transition Transition) Configuration {
	conf, ok := from.(*MorphConfiguration)
	if !ok {
		panic("Got wrong configuration type")
	}
	if transition[:2] == "MD" {
		conf = conf.Copy().(*MorphConfiguration)
		lID, lExists := conf.LatticeQueue.Pop()
		lattice := conf.Lattices[lID]
		if !lExists {
			panic("Can't MD, Lattice Queue is empty")
		}
		_, qExists := conf.Queue().Peek()
		if qExists {
			panic("Can't MD, Queue is not empty")
		}
		spelloutNum, err := strconv.Atoi(string(transition[3:]))
		if err != nil {
			panic("Error converting MD transition # to int:\n" + err.Error())
		}
		lattice.GenSpellouts()
		spellout := lattice.Path(spelloutNum)
		token := lattice.Token
		conf.Mappings = append(conf.Mappings, &NLP.Mapping{token, spellout})
		numNodes := len(conf.MorphNodes)
		spelloutLen := len(spellout)
		var id int
		for i, morpheme := range spellout {
			id = spelloutLen - i - 1 + numNodes
			conf.Queue().Push(id)
			conf.MorphNodes = append(conf.MorphNodes, morpheme)
		}
		conf.SetLastTransition(Transition("MD-" + spellout.String()))
		return conf
	} else {
		copyconf := conf.Copy().(*MorphConfiguration)
		copyconf.SimpleConfiguration = *a.ArcEager.Transition(&conf.SimpleConfiguration, transition).(*SimpleConfiguration)
		return copyconf
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
	if !ok {
		panic("Got wrong configuration type")
	}
	_, qExists := conf.Queue().Peek()
	latticeID, lExists := conf.LatticeQueue.Pop()
	lattice := conf.Lattices[latticeID]
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
		a.oracle = Oracle(&ArcEagerMorphOracle{})
		a.ArcEager.AddDefaultOracle()
	}
}

func (a *ArcEagerMorph) Oracle() Oracle {
	return a.oracle
}

type ArcEagerMorphOracle struct {
	ArcEagerOracle
	morphGold []*NLP.Mapping
}

var _ Decision = &ArcEagerMorphOracle{}

func (o *ArcEagerMorphOracle) SetGold(g interface{}) {
	morphGold, ok := g.(NLP.MorphDependencyGraph)
	if !ok {
		panic("Gold is not a morph dependency graph")
	}
	o.morphGold = morphGold.GetMappings()
	o.ArcEagerOracle.SetGold(g)
}

func (o *ArcEagerMorphOracle) Transition(conf Configuration) Transition {
	c := conf.(*MorphConfiguration)
	if o.morphGold == nil {
		panic("Oracle neds gold reference, use SetGold")
	}
	latticeID, lExists := c.LatticeQueue.Peek()
	_, bExists := c.Queue().Peek()
	if lExists && !bExists {
		lattice := c.Lattices[latticeID]
		mapping := o.morphGold[len(c.Mappings)]
		lattice.GenSpellouts()
		pathId, exists := lattice.Spellouts.Find(mapping.Spellout)
		if !exists {
			panic("Oracle can't find oracle spellout in instance lattice")
		}
		return Transition("MD-" + strconv.Itoa(pathId))
	} else {
		return o.ArcEagerOracle.Transition(&c.SimpleConfiguration)
	}
}
