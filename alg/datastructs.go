package alg

import "reflect"

type Index interface {
	Index(int) (int, bool)
}

type Stack interface {
	Index
	Clear()
	Push(int)
	Pop() (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Stack
	Equal(Stack) bool
}

type Queue interface {
	Index
	Clear()
	Push(int)
	Enqueue(int)
	Dequeue() (int, bool)
	Pop() (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Queue
	Equal(Queue) bool
}

type StackArray struct {
	Array []int
}

var _ Stack = &StackArray{}

func (s *StackArray) Equal(other Stack) bool {
	return reflect.DeepEqual(s, other)
}

func (s *StackArray) Clear() {
	s.Array = s.Array[0:0]
}

func (s *StackArray) Push(val int) {
	s.Array = append(s.Array, val)
}

func (s *StackArray) Pop() (int, bool) {
	if s.Size() == 0 {
		return 0, false
	}
	retval := s.Array[len(s.Array)-1]
	s.Array = s.Array[:len(s.Array)-1]
	return retval, true
}

func (s *StackArray) Index(index int) (int, bool) {
	if index >= s.Size() {
		return 0, false
	}
	return s.Array[len(s.Array)-1-index], true
}

func (s *StackArray) Peek() (int, bool) {
	result, exists := s.Index(0)
	return result, exists
}

func (s *StackArray) Size() int {
	return len(s.Array)
}

func (s *StackArray) Copy() Stack {
	newArray := make([]int, len(s.Array), cap(s.Array))
	copy(newArray, s.Array)
	newStack := Stack(&StackArray{newArray})
	return newStack
}

func NewStackArray(size int) *StackArray {
	return &StackArray{make([]int, 0, size)}
}

type QueueSlice struct {
	slice       []int
	hasDequeued bool
}

var _ Queue = &QueueSlice{}

func (q *QueueSlice) Clear() {
	q.slice = q.slice[0:0]
}

func (q *QueueSlice) Equal(other Queue) bool {
	return reflect.DeepEqual(q, other)
}

func (q *QueueSlice) Enqueue(val int) {
	// if q.hasDequeued {
	// 	panic("Can't Enqueue after Dequeue")
	// }
	q.slice = append(q.slice, val)
}

func (q *QueueSlice) Dequeue() (int, bool) {
	if q.Size() == 0 {
		return 0, false
	}
	retval := q.slice[0]
	q.slice = q.slice[1:]
	return retval, true
}

func (q *QueueSlice) Index(index int) (int, bool) {
	if index >= q.Size() {
		return 0, false
	}
	return q.slice[index], true
}

func (q *QueueSlice) Peek() (int, bool) {
	result, exists := q.Index(0)
	return result, exists
}

// Pop is mapped to the top of the queue, so it acts like a dequeue
func (q *QueueSlice) Pop() (int, bool) {
	return q.Dequeue()
}

func (q *QueueSlice) Push(val int) {
	oldSlice := q.slice
	q.slice = make([]int, 1+len(oldSlice), 1+cap(oldSlice))
	q.slice[0] = val
	copy(q.slice[1:], oldSlice)
}

func (q *QueueSlice) Size() int {
	return len(q.slice)
}

func (q *QueueSlice) Copy() Queue {
	newSlice := make([]int, len(q.slice), cap(q.slice))
	copy(newSlice, q.slice)
	return &QueueSlice{newSlice, q.hasDequeued}
}

func NewQueueSlice(size int) *QueueSlice {
	return &QueueSlice{make([]int, 0, size), false}
}
