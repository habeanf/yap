package Search

import "sync"

// import "log"

type Agenda interface {
	Contains(Candidate) bool
	Len() int
	Clear()
}

type Problem interface{}
type Candidate interface {
	Copy() Candidate
}
type Candidates []Candidate

type Interface interface {
	StartItem(p Problem) Candidates
	Clear(Agenda) Agenda
	Insert(cs chan Candidate, a Agenda) Agenda
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
		i int
	)

	// candidates <- {STARTITEM(problem)}
	candidates := b.StartItem(problem)
	// agenda <- CLEAR(agenda)
	agenda = b.Clear(agenda)
	// loop do
	for {
		// log.Println()
		// log.Println()
		// log.Println("At gold sequence", i)

		var wg sync.WaitGroup
		// for each candidate in candidates
		for i, candidate := range candidates {
			wg.Add(1)
			go func(ag Agenda, cand Candidate, j int) {
				defer wg.Done()
				// agenda <- INSERT(EXPAND(candidate,problem),agenda)
				agenda = b.Insert(b.Expand(cand, problem, j), ag)
			}(agenda, candidate, i)
			if !b.Concurrent() {
				wg.Wait()
			}
		}
		wg.Wait()

		// if agenda.Len() == 0 {
		// 	// if the agenda is empty, yet the goal is not met
		// 	// we return the previous best result and gold
		// 	// this is also a really bad sign something has gone horribly wrong
		// 	break
		// }
		// for each candidate in candidates
		// for _, candidate := range candidates {
		// 	// agenda <- INSERT(EXPAND(candidate,problem),agenda)
		// 	agenda = b.Insert(b.Expand(candidate, problem), agenda)
		// }

		// best <- TOP(AGENDA)
		best = b.Top(agenda)

		// log.Println("Best:", best)
		// log.Println()
		// log.Println("Agenda:")

		// for i, c := range agenda {
		// 	if i == B {
		// 		log.Println("----- end beam -----")
		// 	}
		// 	log.Println(c)
		// }

		// if GOALTEST(problem,best)
		if b.GoalTest(problem, best) {
			// return best
			break
		}

		// early update
		if earlyUpdate {
			goldValue = goldSequence[i]
			i++
			// if we're on early update and either:
			// a. gold isn't on the agenda
			// b. next gold is
			if !agenda.Contains(goldValue) || i >= len(goldSequence) {
				// log.Println("Early update after", i)
				b.SetEarlyUpdate(i)
				break
			}
		}

		// candidates <- TOP-B(agenda, B)
		candidates = b.TopB(agenda, B)

		// agenda <- CLEAR(agenda)
		agenda = b.Clear(agenda)

		// for i, c := range candidates {
		// 	if i == B {
		// 		log.Println("----- end beam -----")
		// 	}
		// 	log.Println(c)
		// }

	}
	best = best.Copy()
	agenda = b.Clear(agenda)
	return best, goldValue
}
