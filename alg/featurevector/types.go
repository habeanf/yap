package featurevector

// import "fmt"

// import "encoding/gob"

// type GobCodec interface {
// 	gob.GobEncoder
// 	gob.GobDecoder
// }

// type Feature GobCodec

const BASE_SIZE int = 10

var (
	zeroInts   [BASE_SIZE]int
	zeroInt64s [BASE_SIZE]int64
)

type Feature interface{}

type ScoredStore interface {
	Get(transition int) (int64, bool)
	Set(transition int, score int64)
	SetTransitions(transitions []int)
	IncAll(store TransitionScoreStore)
	Inc(transition int, score int64)
	Len() int
	Clear()
	Init()
}

type ArrayStore struct {
	dataArray []int64
}

func (s *ArrayStore) Get(transition int) (int64, bool) {
	if transition > len(s.dataArray) {
		return 0, false
	}
	return s.dataArray[transition], true
}
func (s *ArrayStore) Set(transition int, score int64) {
	if len(s.dataArray) < transition {
		s.dataArray[transition] = score
	}
}

func (s *ArrayStore) SetTransitions(transitions []int) {
	if cap(s.dataArray) < len(transitions) {
		s.dataArray = make([]int64, len(transitions))
	}
}

func (s *ArrayStore) IncAll(store TransitionScoreStore) {
	var val *HistoryValue
	for i, _ := range s.dataArray {
		val = store.GetValue(i)
		if val != nil {
			s.dataArray[i] += val.Value
		}
	}
}

func (s *ArrayStore) Inc(transition int, score int64) {
	if len(s.dataArray) < transition {
		s.dataArray[transition] += score
	}
}

func (s *ArrayStore) Len() int {
	return len(s.dataArray)
}

func (s *ArrayStore) Clear() {
	for i, _ := range s.dataArray {
		s.dataArray[i] = 0
	}
}

func (s *ArrayStore) Init() {
	s.dataArray = make([]int64, 0, 100)
}

type MapStore struct {
	tArray [BASE_SIZE]int
	sArray [BASE_SIZE]int64
	// dataMap map[int]int64
	transitions []int
	scores      []int64
}

func (s *MapStore) Get(transition int) (int64, bool) {
	// if len(s.scores) != len(s.transitions) {
	// 	panic(fmt.Sprintf("Got different lengths: scores %v vs transitions %v", len(s.scores), len(s.transitions)))
	// }
	for i, val := range s.transitions {
		if val == transition {
			return s.scores[i], true
		}
	}
	return 0, false
	// retVal, exists := s.dataMap[transition]
	// return retVal, exists
}
func (s *MapStore) Set(transition int, score int64) {
	for i, val := range s.transitions {
		if val == transition {
			s.scores[i] = score
		}
	}
	s.scores = append(s.scores, score)
	s.transitions = append(s.transitions, transition)
	// s.dataMap[transition] = score
}

func (s *MapStore) SetTransitions(transitions []int) {
	if cap(s.transitions) < len(transitions) {
		s.transitions = make([]int, len(transitions))
		s.scores = make([]int64, len(transitions))
	} else {
		s.transitions = s.tArray[:len(transitions)]
		s.scores = s.sArray[:len(transitions)]
	}
	copy(s.transitions, transitions)
}
func (s *MapStore) Inc(transition int, score int64) {
	// if len(s.scores) != len(s.transitions) {
	// 	panic(fmt.Sprintf("Got different lengths: scores %v vs transitions %v", len(s.scores), len(s.transitions)))
	// }
	for i, val := range s.transitions {
		if val == transition {
			s.scores[i] += score
		}
	}
	// if cur, exists := s.dataMap[transition]; exists {
	// 	s.dataMap[transition] = cur + score
	// } else {
	// 	s.Set(transition, score)
	// }
}

func (s *MapStore) IncAll(store TransitionScoreStore) {
	var val *HistoryValue
	for i, transition := range s.transitions {
		val = store.GetValue(transition)
		if val != nil {
			s.scores[i] += val.Value
		}
	}
}

func (s *MapStore) Len() int {
	return len(s.scores)
}

func (s *MapStore) Clear() {
	copy(s.tArray[0:BASE_SIZE], zeroInts[:BASE_SIZE])
	copy(s.sArray[0:BASE_SIZE], zeroInt64s[:BASE_SIZE])
	s.transitions = s.tArray[0:0]
	s.scores = s.sArray[0:0]
	// s.dataMap = make(map[int]int64, 5)
}

func (s *MapStore) Init() {
	// s.transitions = make([]int, 0, 5)
	// s.scores = make([]int64, 0, 5)
	s.transitions = s.tArray[0:0]
	s.scores = s.sArray[0:0]
}

type HybridStore struct {
	cutoff int
	ArrayStore
	MapStore
}

func (s *HybridStore) Get(transition int) (int64, bool) {
	if transition < s.cutoff {
		return s.ArrayStore.Get(transition)
	} else {
		return s.MapStore.Get(transition)
	}
}
func (s *HybridStore) Set(transition int, score int64) {
	if transition < s.cutoff {
		s.ArrayStore.Set(transition, score)
	} else {
		s.MapStore.Set(transition, score)
	}
}

func (s *HybridStore) SetTransitions(transitions []int) {

	arrayTransitions := make([]int, 0, len(transitions))
	mapTransitions := make([]int, 0, 5)
	for _, transition := range transitions {
		if transition < s.cutoff {
			arrayTransitions = append(arrayTransitions, transition)
		} else {
			mapTransitions = append(mapTransitions, transition)
		}
	}
	s.ArrayStore.SetTransitions(arrayTransitions)
	s.MapStore.SetTransitions(mapTransitions)
}

func (s *HybridStore) Inc(transition int, score int64) {
	if transition < s.cutoff {
		s.ArrayStore.Inc(transition, score)
	} else {
		s.MapStore.Inc(transition, score)
	}
}

func (s *HybridStore) IncAll(store TransitionScoreStore) {
	s.ArrayStore.IncAll(store)
	s.MapStore.IncAll(store)
}
func (s *HybridStore) Len() int {
	return s.ArrayStore.Len() + s.MapStore.Len()
}

func (s *HybridStore) Clear() {
	s.ArrayStore.Clear()
	s.MapStore.Clear()
}

func (s *HybridStore) Init() {
	s.ArrayStore.Init()
	s.MapStore.Init()
}

var (
	_ ScoredStore = &ArrayStore{}
	_ ScoredStore = &MapStore{}
	_ ScoredStore = &HybridStore{}
)

func MakeScoredStore() interface{} {
	s := &MapStore{}
	s.Init()
	return s
}
