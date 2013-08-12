package Transition

import (
	. "chukuparser/Algorithm/Transition"
	. "chukuparser/NLP"
	"reflect"
	"sort"
	"strings"
)

type StackArray struct {
	array []int
}

var _ Stack = &StackArray{}

func (s *StackArray) Equal(other Stack) bool {
	return reflect.DeepEqual(s, other)
}

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
	newArray := make([]int, len(s.array), cap(s.array))
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
var _ sort.Interface = &ArcSetSimple{}

func (s *ArcSetSimple) Less(i, j int) bool {
	if s.arcset[i].GetHead() < s.arcset[j].GetHead() {
		return true
	}
	if s.arcset[i].GetHead() == s.arcset[j].GetHead() {
		return s.arcset[i].GetModifier() < s.arcset[j].GetModifier()
	}
	return false
}

func (s *ArcSetSimple) Swap(i, j int) {
	s.arcset[i], s.arcset[j] = s.arcset[j], s.arcset[i]
}

func (s *ArcSetSimple) Len() int {
	return s.Size()
}

func (s *ArcSetSimple) ValueComp(i, j int, other *ArcSetSimple) int {
	left := s.arcset[i]
	right := other.arcset[j]
	if reflect.DeepEqual(left, right) {
		return 0
	}
	if left.GetModifier() < right.GetModifier() {
		return 1
	}
	return -1
}

func (s *ArcSetSimple) Equal(other ArcSet) bool {
	if s.Size() == 0 && other.Size() == 0 {
		return true
	}
	copyThis := s.Copy().(*ArcSetSimple)
	copyOther := other.Copy().(*ArcSetSimple)
	sort.Sort(copyThis)
	sort.Sort(copyOther)
	return reflect.DeepEqual(copyThis, copyOther)
}

func (s *ArcSetSimple) Sorted() *ArcSetSimple {
	copyThis := s.Copy().(*ArcSetSimple)
	sort.Sort(copyThis)
	return copyThis
}

func (s *ArcSetSimple) Diff(other ArcSet) (ArcSet, ArcSet) {
	copyThis := s.Copy().(*ArcSetSimple)
	copyOther := other.Copy().(*ArcSetSimple)
	sort.Sort(copyThis)
	sort.Sort(copyOther)

	leftOnly := NewArcSetSimple(copyThis.Len())
	rightOnly := NewArcSetSimple(copyOther.Len())
	i, j := 0, 0
	for i < copyThis.Len() && j < copyOther.Len() {
		comp := copyThis.ValueComp(i, j, copyOther)
		switch {
		case comp == 0:
			i++
			j++
		case comp < 0:
			leftOnly.Add(copyThis.arcset[i])
			i++
		case comp > 0:
			rightOnly.Add(copyOther.arcset[j])
			j++
		}
	}
	return leftOnly, rightOnly
}

func (s *ArcSetSimple) Copy() ArcSet {
	newArcs := make([]LabeledDepArc, len(s.arcset), cap(s.arcset))
	copy(newArcs, s.arcset)
	return ArcSet(&ArcSetSimple{newArcs})
}

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

func (s *ArcSetSimple) String() string {
	arcs := make([]string, s.Size())
	for i, arc := range s.arcset {
		arcs[i] = arc.String()
	}
	return strings.Join(arcs, "\n")
}

func NewArcSetSimple(size int) *ArcSetSimple {
	return &ArcSetSimple{make([]LabeledDepArc, 0, size)}
}

func NewArcSetSimpleFromGraph(graph LabeledDependencyGraph) *ArcSetSimple {
	arcSet := NewArcSetSimple(graph.NumberOfEdges())
	for _, edgeNum := range graph.GetEdges() {
		arc := graph.GetLabeledArc(edgeNum)
		arcSet.Add(arc)
	}
	return arcSet
}
