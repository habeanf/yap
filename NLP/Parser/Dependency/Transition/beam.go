package Transition

import (
	"chukuparser/Algorithm/FeatureVector"
	"chukuparser/Algorithm/Heap"
	"chukuparser/Algorithm/Perceptron"
	BeamSearch "chukuparser/Algorithm/Search"
	"chukuparser/Algorithm/Transition"
	TransitionModel "chukuparser/Algorithm/Transition/Model"
	"chukuparser/NLP/Parser/Dependency"
	NLP "chukuparser/NLP/Types"
	"container/heap"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var allOut bool = true

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
	firstCandidate := &ScoredConfiguration{c, 0.0, 0, nil, 0, 0, true}
	firstCandidates[0] = firstCandidate
	if allOut {
		// log.Println("\t\tSpace left on Agenda, current size: 0")
		// log.Println("\t\tPushed onto Agenda", firstCandidate.Transition, "score", firstCandidate.score)
		// log.Println("\t\tAgenda post push 0:0 , ")
	}
	return firstCandidates
}

func (b *Beam) getMaxSize() int {
	return b.Base.Graph().NumberOfNodes() * 2
}

func (b *Beam) Clear(agenda BeamSearch.Agenda) BeamSearch.Agenda {
	start := time.Now()
	if agenda == nil {
		newAgenda := NewAgenda(b.Size)
		// newAgenda.HeapReverse = true
		agenda = newAgenda
	} else {
		agenda.Clear()
	}
	b.DurClearing += time.Since(start)
	return agenda
}

func (b *Beam) Insert(cs chan BeamSearch.Candidate, a BeamSearch.Agenda) []BeamSearch.Candidate { //BeamSearch.Agenda {
	var (
		tempAgendaSize int
	)
	if b.ShortTempAgenda {
		tempAgendaSize = b.Size
	} else {
		tempAgendaSize = b.estimatedTransitions()
	}
	tempAgenda := NewAgenda(tempAgendaSize)
	tempAgendaHeap := heap.Interface(tempAgenda)
	// tempAgenda.HeapReverse = true
	heap.Init(tempAgendaHeap)
	for c := range cs {
		currentScoredConf := c.(*ScoredConfiguration)
		if b.ShortTempAgenda && tempAgenda.Len() == b.Size {
			// if the temp. agenda is the size of the beam
			// there is no reason to add a new one if we can prune
			// some in the beam's Insert function
			if tempAgenda.Peek().score > currentScoredConf.score {
				// log.Println("\t\tNot pushed onto Beam", currentScoredConf.Transition)
				// if the current score has a worse score than the
				// worst one in the temporary agenda, there is no point
				// to adding it
				continue
			} else {
				// log.Println("\t\tPopped", tempAgenda.Confs[0].Transition, "from beam")
				Heap.Pop(tempAgendaHeap)
			}
		}
		// log.Println("\t\tPushed onto Beam", currentScoredConf.Transition)
		heap.Push(tempAgendaHeap, currentScoredConf)
		// heaping += time.Since(lastMem)
	}
	// lastMem = time.Now()
	// agenda := a.(*Agenda)
	// agenda.Lock()
	// agenda.HeapReverse = true
	// for _, beamAction := range tempAgenda.Confs {
	// 	agenda.AddCandidate(beamAction)
	// }
	// // agenda.Confs = append(agenda.Confs, tempAgenda.Confs...)
	// agenda.Unlock()
	// agending += time.Since(lastMem)

	// return agenda
	retval := make([]BeamSearch.Candidate, len(tempAgenda.Confs))
	for i, cand := range tempAgenda.Confs {
		retval[i] = cand
	}
	return retval
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
	scores := make([]float64, 0, b.estimatedTransitions())
	scorer := b.Model.TransitionModel().(*TransitionModel.AvgMatrixSparse)
	scorer.SetTransitionScores(feats, &scores)
	go func(currentConf DependencyConfiguration, candidateChan chan BeamSearch.Candidate) {
		var (
			transNum int
			score    float64
			score1   float64
		)
		// if allOut {
		// 	log.Println("\tExpanding candidate", candidateNum+1, "last transition", currentConf.GetLastTransition(), "score", candidate.score)
		// 	log.Println("\tCandidate:", candidate.C)
		// }
		for transition := range b.TransFunc.YieldTransitions(currentConf.Conf()) {
			score1 = b.Model.TransitionModel().TransitionScore(transition, feats)
			if int(transition) < len(scores) {
				score = scores[int(transition)]
			} else {
				score = 0.0
			}
			if score != score1 {
				panic(fmt.Sprintf("Got different score for transition %v: %v vs %v", transition, score, score1))
			}
			// score = b.Model.TransitionModel().TransitionScore(transition, feats)
			// log.Printf("\t\twith transition/score %d/%v\n", transition, candidate.Score+score)
			// at this point, the candidate has it's *previous* score
			// insert will do compute newConf's features and model score
			// this is done to allow for maximum concurrency
			// where candidates are created while others are being scored before
			// adding into the agenda
			candidateChan <- &ScoredConfiguration{currentConf, transition, candidate.score + score, newFeatList, candidateNum, transNum, false}

			transNum++
		}
		close(candidateChan)
	}(conf, retChan)
	b.DurExpanding += time.Since(start)
	return retChan
}

func (b *Beam) Top(a BeamSearch.Agenda) BeamSearch.Candidate {
	// start := time.Now()
	agenda := a.(*Agenda)
	if agenda.Len() == 0 {
		panic("Got empty agenda!")
	}
	// agendaHeap := heap.Interface(agenda)
	// agenda.HeapReverse = true
	// // heapify agenda
	// heap.Init(agendaHeap)
	// peeking into an initialized (heapified) array
	if len(agenda.Confs) == 0 {
		panic("Got empty agenda")
	}
	var bestCandidate *ScoredConfiguration
	for _, candidate := range agenda.Confs {
		if bestCandidate == nil || candidate.Score() > bestCandidate.Score() {
			bestCandidate = candidate
		}
	}
	bestCandidate.Expand(b.TransFunc)
	// log.Println("Beam's Best:\n", best)
	// sort.Sort(agendaHeap)
	// b.DurTop += time.Since(start)
	return bestCandidate
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
	agenda := a.(*Agenda).Confs
	candidates := make([]BeamSearch.Candidate, len(agenda))
	for i, candidate := range agenda {
		candidates[i] = candidate
	}
	// assume agenda heap is already size of beam
	// agendaHeap := a.(heap.Interface)
	// heap.Init(agendaHeap)
	// for i := 0; i < B; i++ {
	// 	if len(a.(*Agenda).Confs) > 0 {
	// 		candidate := Heap.Pop(agendaHeap).(BeamSearch.Candidate)
	// 		candidates = append(candidates, candidate)
	// 	} else {
	// 		break
	// 	}
	// }

	// concurrent expansion
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

	// TODO: this should be done once before training
	// abstract casting >:-[
	rawGoldSequence := goldInstance.Decoded().(Transition.Configuration).GetSequence()

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
		parsedFeatures = beamScored.Features
		beamLastFeatures := b.FeatExtractor.Features(beamScored.C)
		parsedFeatures = &TransitionModel.FeaturesList{beamLastFeatures, beamScored.Transition, beamScored.Features}
		// log.Println("Finding first wrong transition")
		// log.Println("Beam Conf")
		// log.Println(beamScored.C.Conf().GetSequence())
		// log.Println("Gold Conf")
		// log.Println(goldScored.C.Conf().GetSequence())
		parsedSeq, goldSeq := beamScored.C.Conf().GetSequence(), goldScored.C.Conf().GetSequence()
		var i int
		for i = len(parsedSeq) - 1; i >= 0; i-- {
			// log.Println("At transition", i, "of", len(parsedSeq)-1)
			// log.Println(parsedSeq[i])
			// log.Println(goldSeq[i])
			if parsedSeq[i].GetLastTransition() != goldSeq[i].GetLastTransition() {
				break
			}
		}
		// log.Println("Found", i)

		// log.Println("Rewinding")
		curBeamConf, curGoldConf := beamScored.C, goldScored.C
		curBeamFeatures, curGoldFeatures := parsedFeatures, goldFeatures
		for j := 0; j <= i; j++ {
			// log.Println("At reverse transition", j)
			// log.Println(curBeamConf)
			// log.Println("\tFirst 6 features")
			// for k := 0; k < 6; k++ {
			// 	feat := b.Model.TransitionModel().(*TransitionModel.AvgMatrixSparse).Formatters[k]
			// 	log.Println("\t\t", feat, "=", feat.Format(curBeamFeatures.Previous.Features[k]))
			// }
			// log.Println(curGoldConf)
			// log.Println("\tFirst 6 features")
			// for k := 0; k < 6; k++ {
			// 	feat := b.Model.TransitionModel().(*TransitionModel.AvgMatrixSparse).Formatters[k]
			// 	log.Println("\t\t", feat, "=", feat.Format(curGoldFeatures.Previous.Features[k]))
			// }
			curBeamConf = curBeamConf.Previous()
			curGoldConf = curGoldConf.Previous()
			curBeamFeatures = curBeamFeatures.Previous
			curGoldFeatures = curGoldFeatures.Previous
		}
		if curBeamFeatures != nil {
			curBeamFeatures.Previous = nil
		}
		if curGoldFeatures != nil {
			curGoldFeatures.Previous = nil
		}
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
	score      float64
	Features   *TransitionModel.FeaturesList

	CandidateNum, TransNum int
	Expanded               bool
}

var _ BeamSearch.Candidate = &ScoredConfiguration{}

func (s *ScoredConfiguration) Score() float64 {
	return s.score
}

func (s *ScoredConfiguration) Equal(otherEq BeamSearch.Candidate) bool {
	other := otherEq.(*ScoredConfiguration)
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
	newCand := &ScoredConfiguration{s.C, s.Transition, s.score, s.Features, s.CandidateNum, s.TransNum, true}
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
	BeamSize    int
	Confs       []*ScoredConfiguration
}

func (a *Agenda) String() string {
	retval := make([]string, len(a.Confs))
	for i, conf := range a.Confs {
		retval[i] = fmt.Sprintf("%v:%v", conf.Transition, conf.score)
	}
	return strings.Join(retval, ",")
}

func (a *Agenda) AddCandidates(cs []BeamSearch.Candidate) {
	for _, c := range cs {
		a.AddCandidate(c)
	}
}

func (a *Agenda) AddCandidate(c BeamSearch.Candidate) {
	scored := c.(*ScoredConfiguration)
	if len(a.Confs) < a.BeamSize {
		if allOut {
			// log.Println("\t\tSpace left on Agenda, current size:", len(a.Confs))
			// if len(a.Confs) > 0 {
			// 	log.Println("\t\tFront was:", a.Confs[0].Transition, "score", a.Confs[0].Score())
			// }
		}
		heap.Push(a, scored)
		if allOut {
			// log.Println("\t\tPushed onto Agenda", scored.Transition, "score", scored.score)
			// log.Println("\t\tAgenda post push", a.ConfStr(), ", ")
		}
		return
	}
	peekScore := a.Peek()
	if !(peekScore.score < scored.score) {
		if allOut {
			// log.Println("\t\tNot pushed onto Agenda", scored.Transition, "score", scored.score)
			// log.Println("\t\tKeeping Current", peekScore.Transition, "score", peekScore.score)
		}
		return
	}

	if allOut {
		// log.Println("\t\tAgenda pre pop", a.ConfStr(), ", ")
	}
	// popped := Heap.Pop(a).(*ScoredConfiguration)
	if allOut {
		// log.Println("\t\tPopped off Agenda", popped.Transition, "score", popped.score)
		// log.Println("\t\tAgenda post pop", a.ConfStr(), ", ")
	}
	_ = Heap.Pop(a).(*ScoredConfiguration)
	heap.Push(a, scored)
	if allOut {
		// log.Println("\t\tPushed onto Agenda", scored.Transition, "score", scored.score)
		// log.Println("\t\tAgenda post push", a.ConfStr(), ", ")
	}
}

func (a *Agenda) ConfStr() string {
	retval := make([]string, len(a.Confs))
	for i, val := range a.Confs {
		retval[i] = fmt.Sprintf("%v:%v", val.Transition, val.Score())
	}
	return strings.Join(retval, " , ")
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

func (a *Agenda) Peek() *ScoredConfiguration {
	return a.Confs[0]
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
		// for _, candidate := range a.Confs {
		// 	candidate.Clear()
		// 	candidate = nil
		// }
		a.Confs = a.Confs[0:0]
	}
}

var _ BeamSearch.Agenda = &Agenda{}
var _ heap.Interface = &Agenda{}

func NewAgenda(size int) *Agenda {
	newAgenda := new(Agenda)
	newAgenda.BeamSize = size
	newAgenda.Confs = make([]*ScoredConfiguration, 0, size)
	return newAgenda
}

func CompareConf(confA, confB *ScoredConfiguration, reverse bool) bool {
	// less in reverse, we want the highest scoring to be first in the heap
	// if reverse {
	// 	return confA.score > confB.score
	// }
	// return confA.score < confB.score // less in reverse, we want the highest scoring to be first in the heap
	var retval bool
	if reverse {
		return confA.score > confB.score
		// if confA.score > confB.score {
		// 	retval = true
		// }
		// if confA.score == confB.score {
		// 	if confA.CandidateNum > confB.CandidateNum {
		// 		retval = true
		// 	}
		// 	if confA.CandidateNum == confB.CandidateNum {
		// 		if confA.TransNum > confB.TransNum {
		// 			retval = true
		// 		}
		// 	}
		// }
	} else {
		return confA.score < confB.score
		// if confA.score < confB.score {
		// 	retval = true
		// }
		// if confA.score == confB.score {
		// 	if confA.CandidateNum > confB.CandidateNum {
		// 		retval = true
		// 	}
		// 	if confA.CandidateNum == confB.CandidateNum {
		// 		if confA.TransNum > confB.TransNum {
		// 			retval = true
		// 		}
		// 	}
		// }
	}
	// if reverse {
	return retval
	// } else {
	// 	return !retval
	// }
}
