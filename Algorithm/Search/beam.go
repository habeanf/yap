package Search

import (
	// "log"
	"sync"
)

type Agenda interface {
	AddCandidates([]Candidate)
	Contains(Candidate) bool
	Len() int
	Clear()
}

type Problem interface{}
type Candidate interface {
	Copy() Candidate
	Equal(Candidate) bool
	Score() float64
}
type Candidates []Candidate

type Interface interface {
	StartItem(p Problem) Candidates
	Clear(Agenda) Agenda
	Insert(cs chan Candidate, a Agenda) []Candidate //Agenda
	Expand(c Candidate, p Problem, candidateNum int) chan Candidate
	Top(a Agenda) Candidate
	GoalTest(p Problem, c Candidate) bool
	TopB(a Agenda, B int) Candidates
	Concurrent() bool
	SetEarlyUpdate(int)
}

func Search(b Interface, problem Problem, B int) Candidate {
	candidate, _ := search(b, problem, B, 1, false, nil)
	return candidate
}

func SearchEarlyUpdate(b Interface, problem Problem, B int, goldSequence []Candidate) (Candidate, Candidate) {
	return search(b, problem, B, 1, true, goldSequence)
}

func search(b Interface, problem Problem, B, topK int, earlyUpdate bool, goldSequence []Candidate) (Candidate, Candidate) {
	var (
		goldValue Candidate
		best      Candidate
		agenda    Agenda
		// for early update
		i                 int
		goldExists        bool
		bestBeamCandidate Candidate
		bestScore         float64
	)
	tempAgendas := make([][]Candidate, 0, B)

	// candidates <- {STARTITEM(problem)}
	candidates := b.StartItem(problem)
	bestBeamCandidate = candidates[0]
	// agenda <- CLEAR(agenda)
	agenda = b.Clear(agenda)
	if earlyUpdate {
		goldValue = goldSequence[0]
	}
	// loop do
	for {
		// log.Println()
		// log.Println()
		// log.Println("At gold sequence", i)

		// early update
		if earlyUpdate {
			goldExists = false
			bestBeamCandidate = nil
		}

		tempAgendas = tempAgendas[0:0]
		var wg sync.WaitGroup
		if len(candidates) > cap(tempAgendas) {
			panic("Should not have more candidates than the capacity of the tempAgenda")
		}
		// for each candidate in candidates
		for i, candidate := range candidates {
			if earlyUpdate {
				if bestBeamCandidate == nil || candidate.Score() > bestScore {
					bestScore = candidate.Score()
					bestBeamCandidate = candidate
					// log.Println("Candidate is best")
				} else {
					// log.Println("Candidate is not best")
				}
				if candidate.Equal(goldValue) {
					goldExists = true
					// log.Println("Candidate is gold")
				}
			}
			tempAgendas = append(tempAgendas, nil)
			wg.Add(1)
			go func(ag Agenda, cand Candidate, j int) {
				defer wg.Done()
				// agenda <- INSERT(EXPAND(candidate,problem),agenda)
				// agenda = b.Insert(b.Expand(candidate, problem, i), agenda)
				tempAgendas[j] = b.Insert(b.Expand(cand, problem, j), ag)
			}(agenda, candidate, i)
			if !b.Concurrent() {
				wg.Wait()
			}
		}
		wg.Wait()

		for _, tempCandidates := range tempAgendas {
			agenda.AddCandidates(tempCandidates)
		}

		// early update
		if earlyUpdate {
			if !goldExists {
				b.SetEarlyUpdate(i)
				if bestBeamCandidate == nil {
					panic("Best Beam Candidate is nil")
				}
				best = bestBeamCandidate
				break
			} else {
				i++
				goldValue = goldSequence[i]
			}
		}

		// best <- TOP(AGENDA)
		best = b.Top(agenda)

		// if GOALTEST(problem,best)
		if b.GoalTest(problem, best) {
			// return best
			break
		}

		// candidates <- TOP-B(agenda, B)
		candidates = b.TopB(agenda, B)

		// agenda <- CLEAR(agenda)
		agenda = b.Clear(agenda)

		// log.Println("Next Round", i)
	}
	best = best.Copy()
	agenda = b.Clear(agenda)
	return best, goldValue
}
