package Search

type Agenda interface{}
type Problem interface{}
type Candidate interface{}
type Candidates []*Candidate

type Interface interface {
	StartItem(p *Problem) *Candidates
	Clear() *Agenda
	Insert(cs chan *Candidates, a *Agenda) *Agenda
	Expand(c *Candidate, a *Agenda) chan *Candidates
	Top(a *Agenda) *Candidate
	GoalTest(p *Problem, c *Candidate) bool
	TopB(a *Agenda, B int) *Candidates
}

func Search(b Interface, problem *Problem, B int) *Candidate {
	candidates := b.StartItem(problem)
	for {
		agenda := b.Clear()
		for _, candidate := range *candidates {
			agenda = b.Insert(b.Expand(candidate, agenda), agenda)
		}
		best := b.Top(agenda)
		if b.GoalTest(problem, best) {
			return best
		}
		candidates = b.TopB(agenda, B)
	}
}
