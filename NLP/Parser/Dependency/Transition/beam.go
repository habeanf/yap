package Transition

import (
	"chukuparser/Algorithm/Model/Perceptron"
	BeamSearch "chukuparser/Algorithm/Search"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
	"chukuparser/NLP/Parser/Dependency"
	"container/heap"
	"log"
	"sort"
	"sync"
)

type Beam struct {
	Base             DependencyConfiguration
	TransFunc        Transition.TransitionSystem
	FeatExtractor    Perceptron.FeatureExtractor
	Model            Dependency.ParameterModel
	Size             int
	NumRelations     int
	ReturnModelValue bool
	ReturnSequence   bool
	ReturnWeights    bool
	Log              bool
	ShortTempAgenda  bool
}

var _ BeamSearch.Interface = &Beam{}
var _ Perceptron.EarlyUpdateInstanceDecoder = &Beam{}
var _ Dependency.DependencyParser = &Beam{}

func (b *Beam) StartItem(p BeamSearch.Problem) BeamSearch.Candidates {
	sent, ok := p.(NLP.TaggedSentence)
	if !ok {
		panic("Problem should be an NLP.TaggedSentence")
	}
	if b.Base == nil {
		panic("Set Base to a DependencyConfiguration to parse")
	}
	if b.TransFunc == nil {
		panic("Set Transition to a Transition.TransitionSystem to parse")
	}
	if b.Model == nil {
		panic("Set Model to Dependency.ParameterModel to parse")
	}
	if b.NumRelations == 0 {
		panic("Number of relations not set")
	}
	b.Base.Conf().Init(sent)

	firstCandidates := make([]BeamSearch.Candidate, 1)
	var modelValue Dependency.ParameterModelValue
	if b.ReturnModelValue {
		modelValue = b.Model.NewModelValue()
	}
	firstCandidates[0] = &ScoredConfiguration{b.Base, 0.0, modelValue}

	return firstCandidates
}

func (b *Beam) Clear() BeamSearch.Agenda {
	// beam size * # of transitions
	return NewAgenda(1)
}

func (b *Beam) Insert(cs chan BeamSearch.Candidate, a BeamSearch.Agenda) BeamSearch.Agenda {
	tempAgenda := NewAgenda(b.estimatedTransitions())
	tempAgendaHeap := heap.Interface(tempAgenda)
	heap.Init(tempAgendaHeap)
	for c := range cs {
		currentScoredConf := c.(*ScoredConfiguration)
		conf := currentScoredConf.C
		feats := b.FeatExtractor.Features(conf)
		featsAsWeights := b.Model.ModelValueOnes(feats)
		if b.ReturnModelValue {
			currentScoredConf.ModelValue.Increment(featsAsWeights)
			currentScoredConf.Score = b.Model.WeightedValue(currentScoredConf.ModelValue).Score()
		} else {
			currentScoredConf.Score += b.Model.WeightedValue(featsAsWeights).Score()
		}
		// if b.ShortTempAgenda && tempAgenda.Len() == b.Size {
		// 	// if the temp. agenda is the size of the beam
		// 	// there is no reason to add a new one if we can prune
		// 	// some in the beam's Insert function
		// 	if tempAgenda.confs[0].Score > currentScoredConf.Score {
		// 		// if the current score has a worse score than the
		// 		// worst one in the temporary agenda, there is no point
		// 		// to adding it
		// 		continue
		// 	} else {
		// 		heap.Pop(tempAgendaHeap)
		// 	}
		// }
		heap.Push(tempAgendaHeap, currentScoredConf)
	}
	agenda := a.(*Agenda)
	agenda.Lock()
	agenda.confs = append(agenda.confs, tempAgenda.confs...)
	agenda.Unlock()
	return agenda
}

func (b *Beam) estimatedTransitions() int {
	return b.NumRelations*2 + 2
}

func (b *Beam) Expand(c BeamSearch.Candidate, p BeamSearch.Problem) chan BeamSearch.Candidate {
	var modelValue Dependency.ParameterModelValue
	candidate := c.(*ScoredConfiguration)
	conf := candidate.C
	retChan := make(chan BeamSearch.Candidate, b.estimatedTransitions())
	go func(currentConf DependencyConfiguration, candidateChan chan BeamSearch.Candidate) {
		for transition := range b.TransFunc.YieldTransitions(currentConf.Conf()) {
			newConf := b.TransFunc.Transition(currentConf.Conf(), transition)

			if b.ReturnModelValue {
				modelValue = candidate.ModelValue.Copy()
			}
			// at this point, the candidate has it's *previous* score
			// insert will do compute newConf's features and model score
			// this is done to allow for maximum concurrency
			// where candidates are created while others are being scored before
			// adding into the agenda
			candidateChan <- &ScoredConfiguration{newConf.(DependencyConfiguration), candidate.Score, modelValue}
		}
		close(candidateChan)
	}(conf, retChan)
	return retChan
}

func (b *Beam) Top(a BeamSearch.Agenda) BeamSearch.Candidate {
	agenda := a.(*Agenda)
	agendaHeap := heap.Interface(agenda)
	agenda.HeapReverse = true
	heap.Init(agendaHeap)
	// peeking into an initialized (heapified) array
	best := agenda.confs[0]
	sort.Sort(agendaHeap)
	return best
}

func (b *Beam) GoalTest(p BeamSearch.Problem, c BeamSearch.Candidate) bool {
	conf := c.(*ScoredConfiguration).C
	return conf.Conf().Terminal()
}

func (b *Beam) TopB(a BeamSearch.Agenda, B int) BeamSearch.Candidates {
	candidates := make([]BeamSearch.Candidate, 0, B)
	agendaHeap := a.(heap.Interface)
	// assume agenda heap is already heapified
	heap.Init(agendaHeap)
	for i := 0; i < B; i++ {
		if len(a.(*Agenda).confs) > 0 {
			candidates = append(candidates, heap.Pop(agendaHeap))
		} else {
			break
		}
	}
	return candidates
}

func (b *Beam) Parse(sent NLP.Sentence, constraints Dependency.ConstraintModel, model Dependency.ParameterModel) (NLP.DependencyGraph, interface{}) {
	b.Model = model

	return nil, nil
}

// Perceptron function
func (b *Beam) DecodeEarlyUpdate(goldInstance Perceptron.DecodedInstance, m Perceptron.Model) (Perceptron.DecodedInstance, *Perceptron.SparseWeightVector, *Perceptron.SparseWeightVector) {
	log.Println("Starting decode")
	sent := goldInstance.Instance().(NLP.Sentence)
	b.Model = Dependency.ParameterModel(&PerceptronModel{m.(*Perceptron.LinearPerceptron)})

	// abstract casting >:-[
	rawGoldSequence := goldInstance.Decoded().(Transition.Configuration).GetSequence()

	// drop the first (seq are in reverse) configuration, as it is the initial one
	// which is by definition without a score or features
	rawGoldSequence = rawGoldSequence[:len(rawGoldSequence)-1]

	goldSequence := make([]interface{}, len(rawGoldSequence))
	goldModelValue := b.Model.NewModelValue()
	for i := len(rawGoldSequence) - 1; i >= 0; i-- {
		val := rawGoldSequence[i]
		goldFeat := b.FeatExtractor.Features(val)
		goldAsWeights := b.Model.ModelValueOnes(goldFeat)
		goldModelValue.Increment(goldAsWeights)
		goldSequence[len(rawGoldSequence)-i-1] = &ScoredConfiguration{val.(DependencyConfiguration), goldModelValue.Score(), goldModelValue.Copy()}
	}

	b.ReturnModelValue = true

	log.Println("Begin search..")
	beamResult, goldResult := BeamSearch.SearchConcurrentEarlyUpdate(b, sent, b.Size, goldSequence)
	log.Println("Search ended")

	beamScored := beamResult.(*ScoredConfiguration)
	goldScored := goldResult.(*ScoredConfiguration)
	if b.Log {
		log.Println("Beam Sequence")
		log.Println("\n", beamScored.C.Conf().GetSequence().String())
		log.Println("Gold")
		log.Println("\n", goldScored.C.Conf().GetSequence().String())
	}

	parsedGraph := beamScored.C.Graph()

	parsedWeights := beamScored.ModelValue.(*PerceptronModelValue).vector
	goldWeights := goldScored.ModelValue.(*PerceptronModelValue).vector
	// if b.Log {
	// 	log.Println("Beam Weights")
	// 	log.Println(parsedWeights)
	// 	log.Println("Gold Weights")
	// 	log.Println(goldWeights)
	// }

	return &Perceptron.Decoded{goldInstance.Instance(), parsedGraph}, parsedWeights, goldWeights
}

type ScoredConfiguration struct {
	C          DependencyConfiguration
	Score      float64
	ModelValue Dependency.ParameterModelValue
}

type Agenda struct {
	sync.Mutex
	HeapReverse bool
	confs       []*ScoredConfiguration
}

func (a *Agenda) Len() int {
	return len(a.confs)
}

func (a *Agenda) Less(i, j int) bool {
	scoredI := a.confs[i]
	scoredJ := a.confs[j]
	// less in reverse, we want the highest scoring to be first in the heap
	if a.HeapReverse {
		return scoredI.Score > scoredJ.Score
	}
	return scoredI.Score < scoredJ.Score
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

func (a *Agenda) Contains(goldCandidate BeamSearch.Candidate) bool {
	for _, candidate := range a.confs {
		if candidate.C.Equal(goldCandidate.(*ScoredConfiguration).C) {
			return true
		}
	}
	return false
}

func (a *Agenda) Candidates() BeamSearch.Candidates {
	candidates := make([]BeamSearch.Candidate, len(a.confs))
	for i, val := range a.confs {
		candidates[i] = BeamSearch.Candidate(val)
	}
	return candidates
}

var _ BeamSearch.Agenda = &Agenda{}
var _ heap.Interface = &Agenda{}

func NewAgenda(size int) *Agenda {
	newAgenda := new(Agenda)
	newAgenda.confs = make([]*ScoredConfiguration, 0, size)
	return newAgenda
}
