package util

import (
	"log"
	"runtime"
)

func RangeInt(to int) []int {
	retval := make([]int, to)
	for i := 0; i < to; i++ {
		retval[i] = i
	}
	return retval
}

func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Strcmp(a, b string) int {
	min := len(b)
	if len(a) < len(b) {
		min = len(a)
	}
	diff := 0
	for i := 0; i < min && diff == 0; i++ {
		diff = int(a[i]) - int(b[i])
	}
	if diff == 0 {
		diff = len(a) - len(b)
	}
	return diff
}

func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func LogMemory() {
	s := &runtime.MemStats{}
	runtime.ReadMemStats(s)
	log.Println("*** Memory Info ***")
	log.Println("Bytes Allocated InUse:\t", s.Alloc)
	log.Println("Mallocs:\t\t", s.Mallocs)
	log.Println("Frees:\t\t\t", s.Frees)
	log.Println("Heap Allocated InUse:\t", s.HeapAlloc)
	log.Println("Heap Releases:\t\t", s.HeapReleased)
	log.Println("Heap Objects:\t\t", s.HeapObjects)
	log.Println("Stack Allocated InUse:\t", s.StackInuse)
	log.Println("MSpan In Use:\t\t", s.MSpanInuse)
	log.Println("MCache In Use:\t\t", s.MCacheInuse)
	log.Println("*** ***")
}
