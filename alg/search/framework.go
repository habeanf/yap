package search

import (
	"fmt"
	"log"
	"sync"
)

var AllOut bool = false

type Agenda interface {
	AddCandidates([]Candidate, Candidate) Candidate
	Contains(Candidate) bool
	Len() int
	Clear()
}

type Problem interface{}
type Candidate interface {
	Copy() Candidate
	Equal(Candidate) bool
	Score() int64
}
type Candidates interface {
	Get(int) Candidate
	Len() int
}

type Interface interface {
	StartItem(p Problem) []Candidate
	Clear(Agenda) Agenda
	Insert(cs chan Candidate, a Agenda) []Candidate //Agenda
	Expand(c Candidate, p Problem, candidateNum int) chan Candidate
	Top(a Agenda) Candidate
	Best(a Agenda) Candidate
	GoalTest(p Problem, c Candidate, rounds int) bool
	TopB(a Agenda, B int) []Candidate
	Concurrent() bool
	SetEarlyUpdate(int)

	Name() string
}

func Search(b Interface, problem Problem, B int) Candidate {
	candidate, _ := search(b, problem, B, 1, false, nil)
	return candidate
}

func SearchEarlyUpdate(b Interface, problem Problem, B int, goldSequence Candidates) (Candidate, Candidate) {
	return search(b, problem, B, 1, true, goldSequence)
}

func search(b Interface, problem Problem, B, topK int, earlyUpdate bool, goldSequence Candidates) (Candidate, Candidate) {
	var (
		goldValue Candidate
		best      Candidate
		agenda    Agenda
		// for early update
		i                 int
		goldExists        bool
		bestBeamCandidate Candidate
		resultsReady      chan chan int
	)
	tempAgendas := make([][]Candidate, 0, B)

	// candidates <- {STARTITEM(problem)}
	candidates := b.StartItem(problem)
	bestBeamCandidate = candidates[0]
	// agenda <- CLEAR(agenda)
	agenda = b.Clear(agenda)
	if earlyUpdate {
		goldValue = goldSequence.Get(0)
	}
	// loop do
	for {
		// log.Println()
		// log.Println()
		// log.Println("At gold sequence", i)

		// early update
		if earlyUpdate {
			goldExists, bestBeamCandidate = false, nil
		}

		best = nil
		tempAgendas = tempAgendas[0:0]
		resultsReady = make(chan chan int, B)
		var wg sync.WaitGroup
		if len(candidates) > cap(tempAgendas) {
			panic(fmt.Sprintf("Should not have more candidates than the capacity of the tempAgenda: (%d,%d)\n", len(candidates), cap(tempAgendas)))
		}
		// for each candidate in candidates
		go func() {
			for i, candidate := range candidates {
				tempAgendas = append(tempAgendas, nil)
				readyChan := make(chan int, 1)
				resultsReady <- readyChan
				wg.Add(1)
				go func(ag Agenda, cand Candidate, j int, doneChan chan int) {
					defer wg.Done()
					// agenda <- INSERT(EXPAND(candidate,problem),agenda)
					// tempAgendas[i] = b.Insert(b.Expand(candidate, problem, i), agenda)
					tempAgendas[j] = b.Insert(b.Expand(cand, problem, j), ag)

					doneChan <- j
					close(doneChan)
					// readyChan <- i
					// close(readyChan)
					if !b.Concurrent() {
						best = agenda.AddCandidates(tempAgendas[j], best)
					}
				}(agenda, candidate, i, readyChan)
				if !b.Concurrent() {
					wg.Wait()
					// best = agenda.AddCandidates(tempAgendas[i], best)
				}

				if earlyUpdate {
					if bestBeamCandidate == nil || candidate.Score() > bestBeamCandidate.Score() {
						// bestScore = candidate.Score()
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
				if !b.Concurrent() {
					wg.Wait()
				}
			}
			close(resultsReady)
		}()
		// wg.Wait()

		for readyChan := range resultsReady {
			if b.Concurrent() {
				for tempAgendaId := range readyChan {
					best = agenda.AddCandidates(tempAgendas[tempAgendaId], best)
				}
			} else {
				for _ = range readyChan {
				}
			}
		}

		// for _, tempCandidates := range tempAgendas {
		// 	agenda.AddCandidates(tempCandidates)
		// }
		i++

		// early update
		if earlyUpdate {
			if !goldExists || i >= goldSequence.Len() {
				if AllOut {
					log.Println("EARLY UPDATE")
				}
				b.SetEarlyUpdate(i - 1)
				if bestBeamCandidate == nil {
					panic("Best Beam Candidate is nil")
				}
				best = bestBeamCandidate
				break
			} else {
				goldValue = goldSequence.Get(i)
			}
		}

		// best <- TOP(AGENDA)
		if earlyUpdate {
			best = b.Top(agenda)
		}
		// if GOALTEST(problem,best)
		if b.GoalTest(problem, best, i) {
			if AllOut {
				log.Println("Next Round", i-1)
			}

			// return best
			break
		}

		// candidates <- TOP-B(agenda, B)
		candidates = b.TopB(agenda, B)

		// agenda <- CLEAR(agenda)
		agenda = b.Clear(agenda)

		if AllOut {
			log.Println("Next Round", i-1)
		}
	}
	if !earlyUpdate {
		best = b.Best(agenda)
	}
	best = best.Copy()
	agenda = b.Clear(agenda)
	return best, goldValue
}
