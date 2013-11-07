// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Altered by Amir More (2013) to match libstdc implementation

// Package heap provides heap operations for any type that implements
// heap.Interface. A heap is a tree with the property that each node is the
// minimum-valued node in its subtree.
//
// A heap is a common way to implement a priority queue. To build a priority
// queue, implement the Heap interface with the (negative) priority as the
// ordering for the Less method, so Push adds items while Pop removes the
// highest-priority item from the queue. The Examples include such an
// implementation; the file example_pq_test.go has the complete source.
//
package rlheap

import (
	"container/heap"
	// "log"
)

// var Logging = false

// A heap must be initialized before any of the heap operations
// can be used. Init is idempotent with respect to the heap invariants
// and may be called whenever the heap invariants may have been invalidated.
// Its complexity is O(n) where n = h.Len().
//
func Init(h heap.Interface) {
	// heapify
	n := h.Len()
	for i := n/2 - 1; i >= 0; i-- {
		down(h, i, n)
	}
}

// Push pushes the element x onto the heap. The complexity is
// O(log(n)) where n = h.Len().
//
func Push(h heap.Interface, x interface{}) {
	h.Push(x)
	up(h, h.Len()-1)
}

// Pop removes the minimum element (according to Less) from the heap
// and returns it. The complexity is O(log(n)) where n = h.Len().
// Same as Remove(h, 0).
//
func Pop(h heap.Interface) interface{} {
	// log.Println("Popping")
	n := h.Len() - 1
	h.Swap(0, n)
	down(h, 0, n)
	return h.Pop()
}

// Remove removes the element at index i from the heap.
// The complexity is O(log(n)) where n = h.Len().
//
func Remove(h heap.Interface, i int) interface{} {
	n := h.Len() - 1
	if n != i {
		h.Swap(i, n)
		down(h, i, n)
		up(h, i)
	}
	return h.Pop()
}

func up(h heap.Interface, j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		j = i
	}
}

func down(h heap.Interface, i, n int) {
	for {
		// golang's standard package:
		// j1 := i*2 + 1
		j1 := 2 * (i + 1)
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			// if Logging {
			// 	log.Println("\tBreak 1 (i,j1,n)", i, j1, n)
			// }
			if j1 == n {
				j1 = j1 - 1
				if h.Less(i, j1) == h.Less(j1, i) {
					h.Swap(i, j1)
				}
			}
			break
		}
		j := j1 // left child
		if j2 := j1 - 1; j2 < n && h.Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		less1, less2 := h.Less(i, j), h.Less(j, i)
		if less1 && (less1 != less2) {
			// if Logging {
			// 	log.Println("\tBreak 2")
			// }
			break
		}
		h.Swap(i, j)
		i = j
	}
}

func regularup(h heap.Interface, j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		j = i
	}
}

func regulardown(h heap.Interface, i, n int) {
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !h.Less(j1, j2) {
			j = j2 // = 2*i + 2  // right child
		}
		if !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		i = j
	}
}

func Down(h heap.Interface, i, j int) {
	down(h, i, j)
}

func RegularDown(h heap.Interface, i, j int) {
	regulardown(h, i, j)
}

func Sort(h heap.Interface) {
	for i := h.Len(); i > 1; {
		i--
		// Pop without reslicing
		h.Swap(0, i)
		down(h, 0, i)
	}
	if h.Len() > 1 && h.Less(0, 1) {
		h.Swap(0, 1)
	}
}

func RegularSort(h heap.Interface) {
	for i := h.Len() - 1; i > 1; i-- {
		// Pop without reslicing
		h.Swap(0, i)
		regulardown(h, 0, i)
	}
	if h.Len() > 1 {
		left, right := h.Less(0, 1), h.Less(1, 0)
		if left || !right {
			h.Swap(0, 1)
		}
	}
	// if h.Len() > 1 && h.Less(0, 1) {
	// 	h.Swap(0, 1)
	// }
}
