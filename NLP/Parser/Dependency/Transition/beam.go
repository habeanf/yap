package Transition

import (
	"chukuparser/Algorithm/FeatureVector"
	"chukuparser/Algorithm/Perceptron"
	BeamSearch "chukuparser/Algorithm/Search"
	"chukuparser/Algorithm/Transition"
	TransitionModel "chukuparser/Algorithm/Transition/Model"
	"chukuparser/NLP/Parser/Dependency"
	NLP "chukuparser/NLP/Types"
	"container/heap"
	"log"
	// "sort"
	"sync"
	"time"
)

type Beam struct {
	// main beam functions and parameters
	Base          DependencyConfiguration
	TransFunc     Transition.TransitionSystem
	FeatExtractor Perceptron.FeatureExtractor
	Model         Dependency.TransitionParameterModel
	Size          int
	EarlyUpdateAt int

	// beam parsing variables
	currentBeamSize int

	// parameters for parsing
	// TODO: fold into transition system
	NumRelations int

	// flags
	ReturnModelValue   bool
	ReturnSequence     bool
	ReturnWeights      bool
	ShowConsiderations bool
	ConcurrentExec     bool
	Log                bool
	ShortTempAgenda    bool
	NoRecover          bool

	// used for performance tuning
	lastRoundStart time.Time
	DurTotal       time.Duration
	DurExpanding   time.Duration
	DurInserting   time.Duration
	DurInsertFeat  time.Duration
	DurInsertModl  time.Duration
	DurInsertModA  time.Duration
	DurInsertModB  time.Duration
	DurInsertModC  time.Duration
	DurInsertScrp  time.Duration
	DurInsertScrm  time.Duration
	DurInsertHeap  time.Duration
	DurInsertAgen  time.Duration
	DurInsertInit  time.Duration
	DurTop         time.Duration
	DurTopB        time.Duration
	DurClearing    time.Duration
}

var _ BeamSearch.Interface = &Beam{}
var _ Perceptron.EarlyUpdateInstanceDecoder = &Beam{}
var _ Dependency.DependencyParser = &Beam{}

func (b *Beam) Concurrent() bool {
	return b.ConcurrentExec
}

func (b *Beam) StartItem(p BeamSearch.Problem) BeamSearch.Candidates {
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

	sent, ok := p.(NLP.Sentence)
	if !ok {
		panic("Problem should be an NLP.TaggedSentence")
	}
	c := b.Base.Conf().Copy().(DependencyConfiguration)
	c.Clear()
	c.Conf().Init(sent)

	b.currentBeamSize = 0

	firstCandidates := make([]BeamSearch.Candidate, 1)
	firstCandidates[0] = &ScoredConfiguration{c, 0.0, 0, nil, 0, 0, true}
	return firstCandidates
}

func (b *Beam) getMaxSize() int {
	return b.Base.Graph().NumberOfNodes() * 2
}

func (b *Beam) Clear(agenda BeamSearch.Agenda) BeamSearch.Agenda {
	start := time.Now()
	if agenda == nil {
		agenda = NewAgenda(b.Size * b.Size)
	} else {
		agenda.Clear()
	}
	b.DurClearing += time.Since(start)
	return agenda
}

func (b *Beam) Insert(cs chan BeamSearch.Candidate, a BeamSearch.Agenda) BeamSearch.Agenda {
	var (
		lastMem                      time.Time
		featuring, scoring, modeling time.Duration
		agending, heaping            time.Duration
		initing, scoringModel        time.Duration
		modA, modB, modC             time.Duration
		tempAgendaSize               int
	)
	start := time.Now()
	if b.ShortTempAgenda {
		tempAgendaSize = b.Size
	} else {
		tempAgendaSize = b.estimatedTransitions()
	}
	tempAgenda := NewAgenda(tempAgendaSize)
	tempAgendaHeap := heap.Interface(tempAgenda)
	heap.Init(tempAgendaHeap)
	initing += time.Since(start)
	for c := range cs {
		currentScoredConf := c.(*ScoredConfiguration)
		// lastMem = time.Now()
		// modelScore := b.Model.TransitionModel().TransitionScore(currentScoredConf.Transition, currentScoredConf.Features.Features)
		// scoring += time.Since(lastMem)
		// currentScoredConf.Score += modelScore
		// lastMem = time.Now()
		if b.ShortTempAgenda && tempAgenda.Len() == b.Size {
			// if the temp. agenda is the size of the beam
			// there is no reason to add a new one if we can prune
			// some in the beam's Insert function
			if tempAgenda.Confs[0].Score > currentScoredConf.Score {
				// if the current score has a worse score than the
				// worst one in the temporary agenda, there is no point
				// to adding it
				continue
			} else {
				heap.Pop(tempAgendaHeap)
			}
		}
		heap.Push(tempAgendaHeap, currentScoredConf)
		heaping += time.Since(lastMem)
	}
	lastMem = time.Now()
	agenda := a.(*Agenda)
	agenda.Lock()
	agenda.Confs = append(agenda.Confs, tempAgenda.Confs...)
	agenda.Unlock()
	agending += time.Since(lastMem)

	insertDuration := time.Since(start)
	b.DurInserting += insertDuration
	b.DurInsertFeat += featuring
	b.DurInsertScrp += scoring
	b.DurInsertScrm += scoringModel
	b.DurInsertModl += modeling
	b.DurInsertModA += modA
	b.DurInsertModB += modB
	b.DurInsertModC += modC
	b.DurInsertHeap += heaping
	b.DurInsertAgen += agending
	b.DurInsertInit += initing
	// log.Println("Time featuring (pct):\t", featuring.Nanoseconds(), 100*featuring/insertDuration)
	// log.Println("Time converting (pct):\t", converting.Nanoseconds(), 100*converting/insertDuration)
	// log.Println("Time weighing (pct):\t", weighing.Nanoseconds(), 100*weighing/insertDuration)
	// log.Println("Time scoring (pct):\t", scoring.Nanoseconds(), 100*scoring/insertDuration)
	// log.Println("Time dot scoring (pct):\t", dotScoring.Nanoseconds())
	// log.Println("Inserting Total:", insertDuration)
	// log.Println("Beam State", b.currentBeamSize, "/", b.getMaxSize(), "Ending insert")
	return agenda
}

func (b *Beam) estimatedTransitions() int {
	return b.NumRelations*2 + 2
}

func (b *Beam) Expand(c BeamSearch.Candidate, p BeamSearch.Problem, candidateNum int) chan BeamSearch.Candidate {
	var (
		lastMem   time.Time
		featuring time.Duration
	)
	start := time.Now()
	candidate := c.(*ScoredConfiguration)
	conf := candidate.C
	lastMem = time.Now()
	feats := b.FeatExtractor.Features(conf)
	featuring += time.Since(lastMem)
	var newFeatList *TransitionModel.FeaturesList
	if b.ReturnModelValue {
		newFeatList = &TransitionModel.FeaturesList{feats, conf.GetLastTransition(), candidate.Features}
	} else {
		newFeatList = &TransitionModel.FeaturesList{feats, conf.GetLastTransition(), nil}
	}
	retChan := make(chan BeamSearch.Candidate, b.estimatedTransitions())
	go func(currentConf DependencyConfiguration, candidateChan chan BeamSearch.Candidate) {
		var transNum int
		log.Println("\tExpanding candidate", candidateNum+1, "last transition", currentConf.GetLastTransition())
		for transition := range b.TransFunc.YieldTransitions(currentConf.Conf()) {
			score := b.Model.TransitionModel().TransitionScore(transition, feats)
			// log.Printf("\t\twith transition/score %d/%v\n", transition, score)
			// at this point, the candidate has it's *previous* score
			// insert will do compute newConf's features and model score
			// this is done to allow for maximum concurrency
			// where candidates are created while others are being scored before
			// adding into the agenda
			candidateChan <- &ScoredConfiguration{currentConf, transition, candidate.Score + score, newFeatList, candidateNum, transNum, false}

			transNum++
		}
		close(candidateChan)
	}(conf, retChan)
	b.DurExpanding += time.Since(start)
	return retChan
}

func (b *Beam) Top(a BeamSearch.Agenda) BeamSearch.Candidate {
	start := time.Now()
	agenda := a.(*Agenda)
	if agenda.Len() == 0 {
		panic("Got empty agenda!")
	}
	agendaHeap := heap.Interface(agenda)
	agenda.HeapReverse = true
	// heapify agenda
	heap.Init(agendaHeap)
	// peeking into an initialized (heapified) array
	if len(agenda.Confs) == 0 {
		panic("Got empty agenda")
	}
	best := agenda.Confs[0]
	// log.Println("Beam's Best:\n", best)
	// sort.Sort(agendaHeap)
	best.Expand(b.TransFunc)
	b.DurTop += time.Since(start)
	return best
}

func (b *Beam) SetEarlyUpdate(i int) {
	b.EarlyUpdateAt = i
}

func (b *Beam) GoalTest(p BeamSearch.Problem, c BeamSearch.Candidate) bool {
	conf := c.(*ScoredConfiguration).C
	return conf.Conf().Terminal()
}

func (b *Beam) TopB(a BeamSearch.Agenda, B int) BeamSearch.Candidates {
	start := time.Now()
	candidates := make([]BeamSearch.Candidate, 0, B)
	agendaHeap := a.(heap.Interface)
	// assume agenda heap is already heapified
	heap.Init(agendaHeap)
	for i := 0; i < B; i++ {
		if len(a.(*Agenda).Confs) > 0 {
			candidate := heap.Pop(agendaHeap).(BeamSearch.Candidate)
			candidates = append(candidates, candidate)
		} else {
			break
		}
	}
	// expand concurrently
	var wg sync.WaitGroup
	for _, candidate := range candidates {
		wg.Add(1)
		go func(c BeamSearch.Candidate) {
			c.(*ScoredConfiguration).Expand(b.TransFunc)
			wg.Done()
		}(candidate)
	}
	wg.Wait()
	b.DurTopB += time.Since(start)
	return candidates
}

func (b *Beam) Parse(sent NLP.Sentence, constraints Dependency.ConstraintModel, model Dependency.ParameterModel) (NLP.DependencyGraph, interface{}) {
	start := time.Now()
	prefix := log.Prefix()
	log.SetPrefix("Parsing ")
	b.Model = model.(Dependency.TransitionParameterModel)
	// log.Println("Starting parse")
	beamScored := BeamSearch.Search(b, sent, b.Size).(*ScoredConfiguration)
	// build result parameters
	var resultParams *ParseResultParameters
	if b.ReturnModelValue || b.ReturnSequence {
		resultParams = new(ParseResultParameters)
		if b.ReturnModelValue {
			resultParams.modelValue = beamScored.Features
		}
		if b.ReturnSequence {
			resultParams.Sequence = beamScored.C.Conf().GetSequence()
		}
	}
	configurationAsGraph := beamScored.C.(NLP.DependencyGraph)

	// log.Println("Time Expanding (pct):\t", b.DurExpanding.Nanoseconds(), 100*b.DurExpanding/b.DurTotal)
	// log.Println("Time Inserting (pct):\t", b.DurInserting.Nanoseconds(), 100*b.DurInserting/b.DurTotal)
	// log.Println("Time Inserting-Feat (pct):\t", b.DurInsertFeat.Nanoseconds(), 100*b.DurInsertFeat/b.DurTotal)
	// log.Println("Time Inserting-Scor (pct):\t", b.DurInsertScor.Nanoseconds(), 100*b.DurInsertScor/b.DurTotal)
	// log.Println("Total Time:", b.DurTotal.Nanoseconds())
	log.SetPrefix(prefix)
	b.DurTotal += time.Since(start)
	return configurationAsGraph, resultParams
}

// Perceptron function
func (b *Beam) DecodeEarlyUpdate(goldInstance Perceptron.DecodedInstance, m Perceptron.Model) (Perceptron.DecodedInstance, interface{}, interface{}, int) {
	b.EarlyUpdateAt = -1
	start := time.Now()
	prefix := log.Prefix()
	// log.SetPrefix("Training ")
	// log.Println("Starting decode")
	sent := goldInstance.Instance().(NLP.Sentence)
	transitionModel := m.(TransitionModel.Interface)
	b.Model = Dependency.TransitionParameterModel(&PerceptronModel{transitionModel})

	// abstract casting >:-[
	rawGoldSequence := goldInstance.Decoded().(Transition.Configuration).GetSequence()

	// drop the first (seq are in reverse) configuration, as it is the initial one
	// which is by definition without a score or features
	// rawGoldSequence = rawGoldSequence[:len(rawGoldSequence)-1]

	goldSequence := make([]BeamSearch.Candidate, len(rawGoldSequence))
	var (
		lastFeatures *TransitionModel.FeaturesList
		curFeats     []FeatureVector.Feature
	)
	for i := len(rawGoldSequence) - 1; i >= 0; i-- {
		val := rawGoldSequence[i]
		curFeats = b.FeatExtractor.Features(val)
		lastFeatures = &TransitionModel.FeaturesList{curFeats, val.GetLastTransition(), lastFeatures}
		goldSequence[len(rawGoldSequence)-i-1] = &ScoredConfiguration{val.(DependencyConfiguration), val.GetLastTransition(), 0.0, lastFeatures, 0, 0, true}
	}

	b.ReturnModelValue = true

	// log.Println("Begin search..")
	beamResult, goldResult := BeamSearch.SearchEarlyUpdate(b, sent, b.Size, goldSequence)
	// log.Println("Search ended")

	beamScored := beamResult.(*ScoredConfiguration)
	var (
		goldFeatures, parsedFeatures *TransitionModel.FeaturesList
		goldScored                   *ScoredConfiguration
	)
	if goldResult != nil {
		goldScored = goldResult.(*ScoredConfiguration)
		goldFeatures = goldScored.Features
		parsedFeatures = beamScored.Features.Previous
		// beamLastFeatures := b.FeatExtractor.Features(beamScored.C)
		// parsedFeatures = &TransitionModel.FeaturesList{beamLastFeatures, beamScored.Transition, beamScored.Features}

		curBeamConf, curGoldConf := beamScored.C, goldScored.C
		log.Println("Rolling back to first equal configuration")
		log.Println("Beam Conf")
		log.Println(curBeamConf.Conf().GetSequence())
		log.Println("Gold Conf")
		log.Println(curGoldConf.Conf().GetSequence())
		curBeamFeatures, curGoldFeatures := parsedFeatures, goldFeatures
		var i int
		for curBeamConf != nil && curGoldConf != nil && !curBeamConf.Equal(curGoldConf) {
			log.Println("At transition", i)
			log.Println(curBeamConf)
			log.Println(curGoldConf)
			curBeamConf = curBeamConf.Previous()
			curGoldConf = curGoldConf.Previous()
			curBeamFeatures = curBeamFeatures.Previous
			curGoldFeatures = curGoldFeatures.Previous
			i++
		}
		curBeamFeatures.Previous = nil
		curGoldFeatures.Previous = nil
	}

	// parsedFeatures := beamScored.ModelValue.(*PerceptronModelValue).vector

	if b.Log {
		log.Println("Beam Sequence")
		log.Println("\n", beamScored.C.Conf().GetSequence().String())
		// log.Println("\n", parsedFeatures)
		if goldScored != nil {
			log.Println("Gold")
			log.Println("\n", goldScored.C.Conf().GetSequence().String())
			// log.Println("\n", goldFeatures)
		}
	}

	parsedGraph := beamScored.C.Graph()

	// if b.Log {
	// 	log.Println("Beam Weights")
	// 	log.Println(parsedFeatures)
	// 	log.Println("Gold Weights")
	// 	log.Println(goldFeatures)
	// }

	log.SetPrefix(prefix)
	b.DurTotal += time.Since(start)
	return &Perceptron.Decoded{goldInstance.Instance(), parsedGraph}, parsedFeatures, goldFeatures, b.EarlyUpdateAt
}

func (b *Beam) ClearTiming() {
	b.DurTotal = 0
	b.DurExpanding = 0
	b.DurInserting = 0
	b.DurInsertFeat = 0
	b.DurInsertModl = 0
	b.DurInsertModA = 0
	b.DurInsertModB = 0
	b.DurInsertModC = 0
	b.DurInsertScrp = 0
	b.DurInsertScrm = 0
	b.DurInsertHeap = 0
	b.DurInsertAgen = 0
	b.DurInsertInit = 0
}

type Features struct {
	Features []FeatureVector.Feature
	Previous *Features
}

type ScoredConfiguration struct {
	C          DependencyConfiguration
	Transition Transition.Transition
	Score      float64
	Features   *TransitionModel.FeaturesList

	CandidateNum, TransNum int
	Expanded               bool
}

var _ BeamSearch.Candidate = &ScoredConfiguration{}

func (s *ScoredConfiguration) Equal(other *ScoredConfiguration) bool {
	if s.Expanded {
		if !other.Expanded {
			return other.Equal(s)
		}
		return s.C.Equal(other.C)
	} else {
		if !other.Expanded {
			panic("Can't compare two unexpanded scored configurations")
		}
		return s.Transition == other.C.GetLastTransition() && s.C.Equal(other.C.Previous())
	}
}

func (s *ScoredConfiguration) Clear() {
	s.C.Clear()
	s.C = nil
}

func (s *ScoredConfiguration) Copy() BeamSearch.Candidate {
	newCand := &ScoredConfiguration{s.C, s.Transition, s.Score, s.Features, s.CandidateNum, s.TransNum, true}
	s.C.IncrementPointers()
	return newCand
}

func (s *ScoredConfiguration) Expand(t Transition.TransitionSystem) {
	if !s.Expanded {
		s.C = t.Transition(s.C.(Transition.Configuration), s.Transition).(DependencyConfiguration)
		s.Expanded = true
	}
}

type Agenda struct {
	sync.Mutex
	HeapReverse bool
	Confs       []*ScoredConfiguration
}

func (a *Agenda) Len() int {
	return len(a.Confs)
}

func (a *Agenda) Less(i, j int) bool {
	scoredI := a.Confs[i]
	scoredJ := a.Confs[j]
	return CompareConf(scoredI, scoredJ, a.HeapReverse)
}

func (a *Agenda) Swap(i, j int) {
	a.Confs[i], a.Confs[j] = a.Confs[j], a.Confs[i]
}

func (a *Agenda) Push(x interface{}) {
	scored := x.(*ScoredConfiguration)
	a.Confs = append(a.Confs, scored)
}

func (a *Agenda) Pop() interface{} {
	n := len(a.Confs)
	scored := a.Confs[n-1]
	a.Confs = a.Confs[0 : n-1]
	return scored
}

func (a *Agenda) Contains(goldCandidate BeamSearch.Candidate) bool {
	goldScored := goldCandidate.(*ScoredConfiguration)
	for _, candidate := range a.Confs {
		if candidate.Equal(goldScored) {
			return true
		}
	}
	return false
}

func (a *Agenda) Clear() {
	if a.Confs != nil {
		// nullify all pointers
		for _, candidate := range a.Confs {
			candidate.Clear()
			candidate = nil
		}
		a.Confs = a.Confs[0:0]
	}
}

var _ BeamSearch.Agenda = &Agenda{}
var _ heap.Interface = &Agenda{}

func NewAgenda(size int) *Agenda {
	newAgenda := new(Agenda)
	newAgenda.Confs = make([]*ScoredConfiguration, 0, size)
	return newAgenda
}

func CompareConf(confA, confB *ScoredConfiguration, reverse bool) bool {
	// less in reverse, we want the highest scoring to be first in the heap
	// if reverse {
	// 	return confA.Score > confB.Score
	// }
	// return confA.Score < confB.Score // less in reverse, we want the highest scoring to be first in the heap
	var retval bool
	if reverse {
		if confA.Score > confB.Score {
			retval = true
		}
		if confA.Score == confB.Score {
			if confA.CandidateNum < confB.CandidateNum {
				retval = true
			}
			if confA.CandidateNum == confB.CandidateNum {
				if confA.TransNum < confB.TransNum {
					retval = true
				}
			}
		}
	} else {
		if confA.Score < confB.Score {
			retval = true
		}
		if confA.Score == confB.Score {
			if confA.CandidateNum < confB.CandidateNum {
				retval = true
			}
			if confA.CandidateNum == confB.CandidateNum {
				if confA.TransNum < confB.TransNum {
					retval = true
				}
			}
		}
	}
	// if reverse {
	return retval
	// } else {
	// 	return !retval
	// }
}
