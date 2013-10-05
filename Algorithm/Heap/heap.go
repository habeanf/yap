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
package Heap

import (
	"container/heap"
	// "log"
)

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
		// log.Println("\tj1 := 2*(i+1)")
		j1 := 2 * (i + 1)
		// log.Println("\tj1, i := ", j1, ",", i)
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			// log.Println("\tBreaking (1)")
			break
		}
		j := j1 // left child
		// log.Println("\tj := ", j1)
		if j2 := j1 - 1; j2 < n && h.Less(j2, j1) {
			// log.Println("\tj = j2")
			j = j2 // = 2*i + 2  // right child
		}
		if h.Less(i, j) {
			// 	log.Println("\tBreaking (2)")
			break
		}
		// log.Println("\tSwapping")
		h.Swap(i, j)
		// log.Println("\ti = j")
		i = j
	}
}
