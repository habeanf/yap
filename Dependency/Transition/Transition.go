package Transition

type TransitionSystem interface {
	Transition(from *Configuration, transition string) *Configuration

	TransitionSet() []string
	TransitionTypes() []string

	Projective() bool
	Labeled() bool

	Oracle() *Decision
	SetGold(*Graph)
}

type Decision interface {
	GetTransition(*Configuration) string
}
