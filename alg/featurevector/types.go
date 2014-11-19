package featurevector

// import "encoding/gob"

// type GobCodec interface {
// 	gob.GobEncoder
// 	gob.GobDecoder
// }

// type Feature GobCodec

type Feature interface{}

type ScoredStore interface {
	Get(transition int) (int64, bool)
	Set(transition int, score int64)
	Inc(transition int, score int64)
	Len() int
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

func (s *ArrayStore) Inc(transition int, score int64) {
	if len(s.dataArray) < transition {
		s.dataArray[transition] += score
	} else {
		s.Set(transition, score)
	}
}

func (s *ArrayStore) Len() int {
	return len(s.dataArray)
}

type MapStore struct {
	dataMap map[int]int64
}

func (s *MapStore) Get(transition int) (int64, bool) {
	retVal, exists := s.dataMap[transition]
	return retVal, exists
}
func (s *MapStore) Set(transition int, score int64) {
	s.dataMap[transition] = score
}

func (s *MapStore) Inc(transition int, score int64) {
	if cur, exists := s.dataMap[transition]; exists {
		s.dataMap[transition] = cur + score
	} else {
		s.Set(transition, score)
	}
}

func (s *MapStore) Len() int {
	return len(s.dataMap)
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

func (s *HybridStore) Inc(transition int, score int64) {
	if transition < s.cutoff {
		s.ArrayStore.Inc(transition, score)
	} else {
		s.MapStore.Inc(transition, score)
	}
}

func (s *HybridStore) Len() int {
	return s.ArrayStore.Len() + s.MapStore.Len()
}

func NewStore(size int) ScoredStore {
	s := &MapStore{}
	s.dataMap = make(map[int]int64, 5)
	return s
}

var (
	_ ScoredStore = &ArrayStore{}
	_ ScoredStore = &MapStore{}
	_ ScoredStore = &HybridStore{}
)
