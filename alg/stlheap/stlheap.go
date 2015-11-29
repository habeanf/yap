package stlheap

import (
	"container/heap"
	// "log"
)

type Interface interface {
	heap.Interface
	Copy(i, j int)
	Set(i int, x interface{})
	Get(i int) interface{}
	LessValue(i int, x interface{}) bool
}

func push(h Interface, holeIndex int, x interface{}) {
	parent := (holeIndex - 1) / 2
	for holeIndex > 0 && h.LessValue(parent, x) {
		// log.Println("putting", parent, holeIndex)
		h.Copy(parent, holeIndex)
		holeIndex = parent
		parent = (holeIndex - 1) / 2
	}
	// log.Println("pushing value at", holeIndex)
	h.Set(holeIndex, x)
}

func adjust(h Interface, length int, x interface{}) {
	secondChild, holeIndex := 0, 0
	// log.Println("length", length)
	// log.Println("SecondChild", secondChild)
	for secondChild < (length-1)/2 {
		secondChild = 2 * (secondChild + 1)
		if h.Less(secondChild, secondChild-1) {
			secondChild--
		}
		// log.Println("Set", holeIndex, secondChild)
		h.Swap(holeIndex, secondChild)
		holeIndex = secondChild
		// log.Println("SecondChild", secondChild)
	}
	// log.Println("After loop", secondChild)
	if length&1 == 0 && secondChild == (length-2)/2 {
		secondChild = 2 * (secondChild + 1)
		// log.Println("Set", holeIndex, secondChild-1)
		h.Copy(secondChild-1, holeIndex)
		holeIndex = secondChild - 1
	}
	// log.Println("Last SecondChild", secondChild)
	// log.Println("Pushing at", holeIndex)
	push(h, holeIndex, x)
}

// func Pop(h Interface) interface{} {
// 	n := h.Len() - 1
// 	h.Swap(0, n)
// 	Adjust(h, n)
// 	return h.Pop()
// }

func Sort(h Interface) {
	for i := h.Len() - 1; i > 0; i-- {
		// log.Println(j)
		// Pop without reslicing
		// i := agenda.Len() - 1
		// agenda.Swap(0, i)
		cur := h.Get(i)
		h.Copy(0, i)
		adjust(h, i, cur)
		// rlheap.RegularDown(agenda, 0, i)
		// rlheap.Down(agenda, 0, i)
		// log.Println(agenda.ConfStr())
		// j++
	}
	// for i := h.Len(); i > 1; {
	// 	i--
	// i := h.Len() - 1
	// // Pop without reslicing
	// h.Swap(0, i)
	// Adjust(h, i)
	// }
}
