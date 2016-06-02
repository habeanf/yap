package util

import (
	"log"
	"runtime"
	"sort"
	"strconv"
	. "unicode"
	"unicode/utf8"
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

func Sign(x int) int {
	switch {
	case x > 0:
		return 1
	case x < 0:
		return -1
	default:
		return 0
	}
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

func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func MaxInt(v []int) (retval int) {
	for _, cur := range v {
		if retval < cur {
			retval = cur
		}
	}
	return
}
func NotDigitOrNeg(c rune) bool {
	_, err := strconv.Atoi(string(c))
	return err != nil && c != '-'
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

type TopNStrIntDatum struct {
	S string
	N int
}

type TopNStrIntData []TopNStrIntDatum

func (arr TopNStrIntData) Len() int {
	return len(arr)
}

func (arr TopNStrIntData) Swap(a, b int) {
	arr[a], arr[b] = arr[b], arr[a]
}

func (arr TopNStrIntData) Less(a, b int) bool {
	return arr[a].N > arr[b].N
}

func GetTopNStrInt(m map[string]int, n int) []TopNStrIntDatum {
	data := make(TopNStrIntData, len(m))
	var i int
	for k, v := range m {
		data[i] = TopNStrIntDatum{k, v}
		i++
	}
	sort.Sort(data)
	return data[:Min(len(data), n)]
}

type RuneTester func(r rune) bool

func TestEach(t RuneTester, s string) byte {
	for i, w := 0, 0; i < len(s); i += w {
		runeValue, width := utf8.DecodeRuneInString(s[i:])
		if t(runeValue) {
			return 't'
		}
		w = width
	}
	return 'f'
}

var Testers = []RuneTester{
	IsDigit,
	IsGraphic,
	IsLetter,
	IsLower,
	IsMark,
	IsNumber,
	IsPunct,
	IsSymbol,
	IsTitle,
	IsUpper,
}

func Signature(s string) string {
	indicators := make([]byte, len(Testers))
	for i, t := range Testers {
		indicators[i] = TestEach(t, s)
	}
	return string(indicators)
}

func Prefix(s string, n int) string {
	return s[0:Min(len(s), n)]
}

func Suffix(s string, n int) string {
	return s[Max(len(s)-n, 0):len(s)]
}
