package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/Algorithm/Search"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
	"chukuparser/NLP/Parser/Dependency"
	"container/heap"
	"sync"
)

type Beam struct {
	Base          DependencyConfiguration
	Transition    Transition.TransitionSystem
	FeatExtractor Perceptron.FeatureExtractor
	Model         Dependency.ParameterModel
	Size          int
	NumRelations  int
}

var _ Search.Interface = &Beam{}
var _ Perceptron.EarlyUpdateInstanceDecoder = &Beam{}
var _ Dependency.DependencyParser = &Beam{}

func (b *Beam) StartItem(p Search.Problem) Search.Candidates {
	sent, ok := p.(NLP.TaggedSentence)
	if !ok {
		panic("Problem should be an NLP.TaggedSentence")
	}
	if b.Base == nil {
		panic("Set Base to a DependencyConfiguration to parse")
	}
	if b.Transition == nil {
		panic("Set Transition to a Transition.TransitionSystem to parse")
	}
	if b.Model == nil {
		panic("Set Model to Dependency.ParameterModel to parse")
	}
	b.Base.Conf().Init(sent)

	firstCandidates := make([]Search.Candidate, 1)
	firstCandidates[0] = &ScoredConfiguration{b.Base, 0.0}

	return firstCandidates
}

func (b *Beam) Clear() Search.Agenda {
	newAgenda := new(Agenda)
	// beam size, # of relations * 2 (left/right arcs) + 2 (SH/RE)
	estimatedAgendaSize := b.Size * b.estimatedTransitions()
	newAgenda.confs = make([]*ScoredConfiguration, 0, estimatedAgendaSize)
	return newAgenda
}

func (b *Beam) Insert(cs chan Search.Candidate, a Search.Agenda) Search.Agenda {
	agenda := a.(*Agenda)
	for c := range cs {
		candidate := c.(*ScoredConfiguration)
		conf := candidate.C
		feats := b.FeatExtractor.Features(conf)
		featsAsWeights := b.Model.ModelValue(feats)
		currentScore := b.Model.NewModelValue().ScoreWith(b.Model, featsAsWeights)
		candidate.Score += currentScore
		agenda.Lock()
		agenda.confs = append(agenda.confs, candidate)
		agenda.Unlock()
	}
	return agenda
}

func (b *Beam) estimatedTransitions() int {
	return b.NumRelations*2 + 2
}

func (b *Beam) Expand(c Search.Candidate, p Search.Problem) chan Search.Candidate {
	candidate := c.(*ScoredConfiguration)
	conf := candidate.C
	retChan := make(chan Search.Candidate, b.estimatedTransitions())
	go func(currentConf DependencyConfiguration, candidateChan chan Search.Candidate) {
		for transition := range b.Transition.YieldTransitions(currentConf.Conf()) {
			newConf := b.Transition.Transition(currentConf.Conf(), transition)
			// at this point, the candidate has it's *previous* score
			// insert will do the new computation
			// this is done to allow for maximum concurrency
			// where candidates are created while others are being scored before
			// adding into the agenda
			candidateChan <- &ScoredConfiguration{newConf.(DependencyConfiguration), candidate.Score}
		}
		close(candidateChan)
	}(conf, retChan)
	return retChan
}

func (b *Beam) Top(a Search.Agenda) Search.Candidate {
	agenda := a.(*Agenda)
	agendaHeap := heap.Interface(agenda)
	heap.Init(agendaHeap)
	// peeking into an initalized heap
	best := agenda.confs[0]
	return best
}

func (b *Beam) GoalTest(p Search.Problem, c Search.Candidate) bool {
	conf := c.(DependencyConfiguration)
	return conf.Conf().Terminal()
}

func (b *Beam) TopB(a Search.Agenda, B int) Search.Candidates {
	candidates := make([]Search.Candidate, B)
	agendaHeap := a.(heap.Interface)
	// assume agenda heap is already heapified
	// heap.Init(agendaHeap)
	for i := 0; i < B; i++ {
		candidates[i] = heap.Pop(agendaHeap)
	}
	return candidates
}

func (b *Beam) Parse(sent NLP.Sentence, constraints Dependency.ConstraintModel, model Dependency.ParameterModel) (NLP.DependencyGraph, interface{}) {
	return nil, nil
}

func (b *Beam) ParseOracleEarlyUpdate(sent NLP.Sentence, gold NLP.DependencyGraph, constraints interface{}, model interface{}) (NLP.DependencyGraph, interface{}, interface{}) {
	return nil, nil, nil
}

func (b *Beam) DecodeEarlyUpdate(goldInstance Perceptron.DecodedInstance, m Perceptron.Model) (Perceptron.DecodedInstance, *Perceptron.SparseWeightVector, *Perceptron.SparseWeightVector) {
	return nil, nil, nil
}

type ScoredConfiguration struct {
	C     DependencyConfiguration
	Score float64
}

type Agenda struct {
	sync.Mutex
	confs []*ScoredConfiguration
}

func (a *Agenda) Len() int {
	return len(a.confs)
}

func (a *Agenda) Less(i, j int) bool {
	scoredI := a.confs[i]
	scoredJ := a.confs[j]
	// less in reverse, we want the highest scoring to be first in the heap
	return scoredI.Score > scoredJ.Score
}

func (a *Agenda) Swap(i, j int) {
	a.confs[i], a.confs[j] = a.confs[j], a.confs[i]
}

func (a *Agenda) Push(x interface{}) {
	scored := x.(*ScoredConfiguration)
	a.confs = append(a.confs, scored)
}

func (a *Agenda) Pop() interface{} {
	n := len(a.confs)
	scored := a.confs[n-1]
	a.confs = a.confs[0 : n-1]
	return scored
}

var _ Search.Agenda = &Agenda{}
var _ heap.Interface = &Agenda{}
