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
}

type Queue interface {
	Clear()
	Enqueue(int)
	Dequeue() (int, bool)
	Index(int) (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Queue
}

type ArcSet interface {
	Clear()
	Add(LabeledDepArc)
	Get(LabeledDepArc) []LabeledDepArc
	Size() int
	Last() LabeledDepArc
	Index(int) LabeledDepArc

	Copy() ArcSet
}

type BaseConfiguration interface {
	Configuration
	Stack() Stack
	Queue() Stack
	Arcs() ArcSet
}
