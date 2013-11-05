package stlheap

import (
	"container/heap"
	// "log"
)

type Interface interface {
	heap.Interface
	Copy(i, j int)
	Set(i int, x interface{})
}

func Push(h Interface, x interface{}) {
	xLocation, holeIndex := h.Len(), h.Len()
	h.Push(x)
	parent := (holeIndex - 1) / 2
	for holeIndex > 0 && h.Less(parent, xLocation) {
		h.Swap(holeIndex, parent)
		holeIndex = parent
		parent = (holeIndex - 1) / 2
	}
}

func adjust(h Interface, length int) {
	topIndex, secondChild, holeIndex := 0, 0, 0
	for secondChild < (length-1)/2 {
		secondChild = 2 * (secondChild + 1)
		if h.Less(secondChild, secondChild-1) {
			secondChild--
		}
		h.Swap(holeIndex, secondChild)
		holeIndex = secondChild
	}
	if length&1 == 0 && secondChild == (length-2)/2 {
		secondChild = 2 * (secondChild + 1)
		h.Copy(holeIndex, secondChild-1)
		holeIndex = secondChild - 1
	}

}

func Pop(h Interface) interface{} {
	n := h.Len() - 1
	h.Swap(0, n)
	adjust(h, n)
	return h.Pop()
}

func Sort(h Interface) {

}
