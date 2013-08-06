package Transition

import (
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP"
)

type StackArray struct {
	array []int
}

var _ Stack = &StackArray{}

func (s *StackArray) Clear() {
	s.array = s.array[0:0]
}

func (s *StackArray) Push(val int) {
	s.array = append(s.array, val)
}

func (s *StackArray) Pop() (int, bool) {
	if s.Size() == 0 {
		return 0, false
	}
	retval := s.array[len(s.array)-1]
	s.array = s.array[:len(s.array)-1]
	return retval, true
}

func (s *StackArray) Index(index int) (int, bool) {
	if index >= s.Size() {
		return 0, false
	}
	return s.array[len(s.array)-1-index], true
}

func (s *StackArray) Peek() (int, bool) {
	result, exists := s.Index(0)
	return result, exists
}

func (s *StackArray) Size() int {
	return len(s.array)
}

func (s *StackArray) Copy() Stack {
	newArray := make([]int, len(s.array))
	copy(newArray, s.array)
	newStack := Stack(&StackArray{newArray})
	return newStack
}

func NewStackArray(size int) *StackArray {
	return &StackArray{make([]int, 0, size)}
}

// type QueueSlice struct {
// 	slice       []int
// 	hasDequeued bool
// }

// func (q *QueueSlice) Clear() {
// 	q.slice = q.slice[0:0]
// }

// func (q *QueueSlice) Enqueue(val int) {
// 	if q.hasDequeued {
// 		panic("Can't Enqueue after Dequeue")
// 	}
// 	q.slice = append(q.slice, val)
// }

// func (q *QueueSlice) Dequeue() (int, bool) {
// 	if q.Size() == 0 {
// 		return 0, false
// 	}
// 	retval := q.slice[0]
// 	return retval, true
// }

// func (q *QueueSlice) Index(index int) (int, bool) {
// 	if index >= q.Size() {
// 		return 0, false
// 	}
// 	return q.slice[index], true
// }

// func (q *QueueSlice) Peek() (int, bool) {
// 	result, exists := q.Index(0)
// 	return result, exists
// }

// func (q *QueueSlice) Size() int {
// 	return len(q.slice)
// }

// func (q *QueueSlice) Copy() QueueSlice {
// 	return QueueSlice{q.slice, q.hasDequeued}
// }

// func NewQueueSlice(size int) QueueSlice {
// 	return QueueSlice{make([]int, 0, size), false}
// }

type ArcSetSimple struct {
	arcset []LabeledDepArc
}

var _ ArcSet = &ArcSetSimple{}

func (s *ArcSetSimple) Clear() {
	s.arcset = s.arcset[0:0]
}

func (s *ArcSetSimple) Index(i int) LabeledDepArc {
	if i >= len(s.arcset) {
		return nil
	}
	return s.arcset[i]
}

func (s *ArcSetSimple) Add(arc LabeledDepArc) {
	s.arcset = append(s.arcset, arc)
}

func (s *ArcSetSimple) Get(query LabeledDepArc) []LabeledDepArc {
	var results []LabeledDepArc
	head := query.GetHead()
	modifier := query.GetModifier()
	relation := query.GetRelation()
	for _, arc := range s.arcset {
		if head >= 0 && head != arc.GetHead() {
			continue
		}
		if modifier >= 0 && modifier != arc.GetModifier() {
			continue
		}
		if relation != "" && relation != arc.GetRelation() {
			continue
		}
		results = append(results, arc)
	}
	return results
}

func (s *ArcSetSimple) Size() int {
	return len(s.arcset)
}

func (s *ArcSetSimple) Last() LabeledDepArc {
	if s.Size() == 0 {
		return nil
	}
	return s.arcset[len(s.arcset)-1]
}

func (s *ArcSetSimple) Copy() ArcSet {
	newArray := make([]LabeledDepArc, len(s.arcset))
	copy(newArray, s.arcset)
	return &ArcSetSimple{newArray}
	// return ArcSet(ArcSetSimple{newArray})
}

func NewArcSetSimple(size int) *ArcSetSimple {
	return &ArcSetSimple{make([]LabeledDepArc, size)}
}
