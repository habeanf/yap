package Transition

type HasAttributes interface {
	GetProperty(property string) (string, bool)
}

type Configuration interface {
	HasAttributes

	Init(interface{})
	Terminal() bool

	Copy() *Configuration
	GetSequence() []*Configuration
	SetLastTransition(string)
	String() string
}

type Stack interface {
	Push(int)
	Pop() (int, bool)
	Index(int) (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Stack
}

type Queue interface {
	Enqueue(int)
	Dequeue() (int, bool)
	Index(int) (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Queue
}

type ArcSet interface {
	Add(DepArc)
	Get(DepArc) []*DepArc
	Size() int
	Last() DepArc

	Copy() ArcSet
}

type BaseConfiguration interface {
	Configuration
	Stack() Stack
	Queue() Queue
	Arcs() ArcSet
}
