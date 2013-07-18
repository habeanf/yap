package Transition

type StackArray struct {
	array []int
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
	if index > s.Size() {
		return 0, false
	}
	return s.array[len(s.array)-1-index], true
}

func (s *StackArray) Peek() (int, bool) {
	return s.Index(0)
}

func (s *StackArray) Size() int {
	return len(s.array)
}

func (s *StackArray) Copy() Stack {
	newArray = make([]int, len(s.array))
	copy(newArray, s.array)
	return StackArray{newArray}
}

func NewStackArray() StackArray {
	return StackArray{make([]int)}
}

type QueueSlice struct {
	slice       []int
	hasDequeued bool
}

func (q *QueueSlice) Enqueue(val int) {
	if q.hasDequeued {
		panic("Can't Enqueue after Dequeue")
	}
	q.slice = append(q.slice, val)
}

func (q *QueueSlice) Dequeue() (int, bool) {
	if q.Size() == 0 {
		return 0, false
	}
	retval := q.slice[0]
	q.slice = q.slice[1:]
	q.hasDequeued = true
	return retval, true
}

func (q *QueueSlice) Index(index int) (int, bool) {
	if index >= q.Size() {
		return 0, false
	}
	return q.slice[index], true
}

func (q *QueueSlice) Peek() (int, bool) {
	return q.Index(0)
}

func (q *QueueSlice) Size() int {
	return len(q.slice)
}

func (q *QueueSlice) Copy() QueueSlice {
	return QueueSlice{q.slice, q.hasDequeued}
}

func NewQueueSlice(size int) QueueSlice {
	return QueueSlice{make([]int, 0, size), false}
}

type ArcSetSimple struct {
	arcset []DepArc
}

func (s *ArcSetSimple) Add(arc DepArc) {
	s.arcset = append(s.arcset, arc)
}

func (s *ArcSetSimple) Get(query DepArc) []*DepArc {
	var results []*DepArc
	for _, arc := range s.arcset {
		if query.Head >= 0 && query.Head != arc.Head {
			continue
		}
		if query.Modifier >= 0 && query.Modifier != arc.Modifier {
			continue
		}
		if query.Relation != "" && query.Relation != arc.Relation {
			continue
		}
		results = append(results, arc)
	}
	return results
}

func (s *ArcSetSimple) Size() int {
	return len(s.arcset)
}

func (s *ArcSetSimple) Last() DepArc {
	if s.Size() == 0 {
		panic("No Arcs in set")
	}
	return s.arcset[len(s.arcset)-1]
}

func (s *ArcSetSimple) Copy() ArcSetSimple {
	newArray := make([]DepArc, len(s.arcset))
	copy(newArray, s.arcset)
	return ArcSetSimple{newArray}
}

func NewArcSetSimple() ArcSet {
	return ArcSetSimple{make([]DepArc)}
}
