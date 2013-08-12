package Util

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
