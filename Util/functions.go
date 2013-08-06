package Util

func RangeInt(to int) []int {
	retval := make([]int, to)
	for i := 0; i < to; i++ {
		retval[i] = i
	}
	return retval
}
