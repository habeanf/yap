package Transition

import . "chukuparser/NLP"

type Stack interface {
	Clear()
	Push(int)
	Pop() (int, bool)
	Index(int) (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Stack
	Equal(Stack) bool
}

type Queue interface {
	Clear()
	Enqueue(int)
	Dequeue() (int, bool)
	Index(int) (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Queue
	Equal(Queue) bool
}

type ArcSet interface {
	Clear()
	Add(LabeledDepArc)
	Get(LabeledDepArc) []LabeledDepArc
	Size() int
	Last() LabeledDepArc
	Index(int) LabeledDepArc

	Copy() ArcSet
	Equal(ArcSet) bool
}

type BaseConfiguration interface {
	Configuration
	Stack() Stack
	Queue() Stack
	Arcs() ArcSet
}
