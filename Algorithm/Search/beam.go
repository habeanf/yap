package Search

import "sync"

type Agenda interface {
	Contains(Candidate) bool
}
type Problem interface{}
type Candidate interface{}
type Candidates []Candidate

type Interface interface {
	StartItem(p Problem) Candidates
	Clear() Agenda
	Insert(cs chan Candidate, a Agenda) Agenda
	Expand(c Candidate, p Problem) chan Candidate
	Top(a Agenda) Candidate
	GoalTest(p Problem, c Candidate) bool
	TopB(a Agenda, B int) Candidates
}

func Search(b Interface, problem Problem, B int) Candidate {
	candidates := b.StartItem(problem)
	for {
		agenda := b.Clear()
		for _, candidate := range candidates {
			agenda = b.Insert(b.Expand(candidate, problem), agenda)
		}
		best := b.Top(agenda)
		if b.GoalTest(problem, best) {
			return best
		}
		candidates = b.TopB(agenda, B)
	}
}

func ConcurrentSearch(b Interface, problem Problem, B int) Candidate {
	candidates := b.StartItem(problem)
	for {
		agenda := b.Clear()
		var wg sync.WaitGroup
		wg.Add(len(candidates))
		for _, candidate := range candidates {
			go func(ag Agenda) {
				defer wg.Done()
				b.Insert(b.Expand(candidate, problem), ag)
			}(agenda)
		}
		wg.Wait()
		best := b.Top(agenda)
		if b.GoalTest(problem, best) {
			return best
		}
		candidates = b.TopB(agenda, B)
	}
}

func SearchEarlyUpdate(b Interface, problem Problem, B int, goldSequence []interface{}) (Candidate, Candidate) {
	candidates := b.StartItem(problem)
	for i, _ := range goldSequence {
		goldValue := goldSequence[len(goldSequence)-i-1]
		agenda := b.Clear()
		for _, candidate := range candidates {
			agenda = b.Insert(b.Expand(candidate, problem), agenda)
		}
		best := b.Top(agenda)
		if !agenda.Contains(goldValue) {
			return best, goldValue
		}
		if b.GoalTest(problem, best) {
			return best, goldValue
		}
		candidates = b.TopB(agenda, B)
	}
	return nil, goldSequence[0]
}
