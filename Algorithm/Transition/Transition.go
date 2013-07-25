package Transition

type Transition string

type Configuration interface {
	Init(interface{})
	Terminal() bool

	Copy() *Configuration
	GetSequence() ConfigurationSequence
	SetLastTransition(Transition)
	String() string
}

type ConfigurationSequence []*Configuration

type TransitionSystem interface {
	Transition(from *Configuration, transition Transition) *Configuration

	TransitionTypes() []Transition

	Oracle() *Decision
}

type Decision interface {
	GetTransition(*Configuration) Transition
}
