package featurevector

import (
	"fmt"
	// "log"
	"strings"
	"sync"
)

type HistoryValue struct {
	sync.Mutex
	Generation     int
	PrevGeneration int
	Value, Total   int64
}

func (h *HistoryValue) Integrate(generation int) {
	h.Value = h.Total + (int64)(generation-h.Generation)*h.Value
}

func (h *HistoryValue) Add(generation int, amount int64) {
	h.Lock()
	defer h.Unlock()
	if h.PrevGeneration < h.Generation {
		h.Total += (int64)(generation-h.Generation) * h.Value
	}
	if h.Generation < generation {
		h.PrevGeneration, h.Generation = h.Generation, generation
	}
	h.Value = h.Value + amount
}

func NewHistoryValue(generation int, value int64) *HistoryValue {
	return &HistoryValue{Generation: generation, Value: value}
}

type LockedArray struct {
	sync.RWMutex
	Vals []*HistoryValue
}

func (l *LockedArray) ExtendFor(generation, transition int) {
	newVals := make([]*HistoryValue, transition+1)
	copy(newVals[0:len(l.Vals)], l.Vals[0:len(l.Vals)])
	l.Vals = newVals
}

func (l *LockedArray) Add(generation, transition int, feature interface{}, amount int64) {
	l.Lock()
	defer l.Unlock()
	if transition < len(l.Vals) {
		if l.Vals[transition] != nil {
			l.Vals[transition].Add(generation, amount)
		} else {
			l.Vals[transition] = NewHistoryValue(generation, amount)
		}
		return
	} else {
		l.ExtendFor(generation, transition)
		if transition >= len(l.Vals) {
			panic("Despite extending, transition >= than Vals")
		}
		l.Vals[transition] = NewHistoryValue(generation, amount)
		return
	}
}

type AvgSparse struct {
	sync.RWMutex
	Vals map[Feature]*LockedArray
}

func (v *AvgSparse) Value(transition int, feature interface{}) int64 {
	transitions, exists := v.Vals[feature]
	if exists && transition < len(transitions.Vals) && transitions.Vals[transition] != nil {
		return transitions.Vals[transition].Value
	}
	return 0.0
}

func (v *AvgSparse) Add(generation, transition int, feature interface{}, amount int64, wg *sync.WaitGroup) {
	v.Lock()
	defer v.Unlock()
	transitions, exists := v.Vals[feature]
	if exists {
		// wg.Add(1)
		go func() {
			transitions.Add(generation, transition, feature, amount)
			wg.Done()
		}()
	} else {
		newTrans := &LockedArray{Vals: make([]*HistoryValue, transition+1)}
		newTrans.Vals[transition] = NewHistoryValue(generation, amount)
		if v.Vals == nil {
			panic("Got nil Vals")
		}
		v.Vals[feature] = newTrans
		wg.Done()
	}
}

func (v *AvgSparse) Integrate(generation int) *AvgSparse {
	for _, val := range v.Vals {
		for _, transition := range val.Vals {
			if transition != nil {
				transition.Integrate(generation)
			}
		}
	}
	return v
}

func (v *AvgSparse) SetScores(feature Feature, scores *[]int64) {
	transitions, exists := v.Vals[feature]
	if exists {
		// log.Println("\t\tSetting scores for feature", feature)
		// log.Println("\t\t\t1. Exists")
		if cap(*scores) < len(transitions.Vals) {
			// log.Println("\t\t\t1.1 Scores array not large enough")
			newscores := make([]int64, len(transitions.Vals))
			// log.Println("\t\t\t1.2 Copying")
			copy(newscores[0:len(transitions.Vals)], (*scores)[0:len(*scores)])
			// log.Println("\t\t\t1.3 Setting pointer")
			*scores = newscores
		}
		// log.Println("\t\t\t2. Iterating", len(transitions), "transitions")
		for i, val := range transitions.Vals {
			if val == nil {
				continue
			}
			// log.Println("\t\t\t\tAt transition", i)
			for len(*scores) <= i {
				// log.Println("\t\t\t\t2.2 extending scores of len", len(*scores), "up to", i)
				*scores = append(*scores, 0)
			}
			// log.Println("\t\t\t\t2.3 incrementing with", val.Value)
			(*scores)[i] += val.Value
		}
		// log.Println("\t\tReturning scores array", *scores)
	}
}

func (v *AvgSparse) UpdateScalarDivide(byValue int64) *AvgSparse {
	if byValue == 0.0 {
		panic("Divide by 0")
	}
	v.RLock()
	defer v.RUnlock()
	for _, val := range v.Vals {
		for _, transition := range val.Vals {
			transition.Value = transition.Value / byValue
		}
	}
	return v
}

func (v *AvgSparse) String() string {
	strs := make([]string, 0, len(v.Vals))
	v.RLock()
	defer v.RUnlock()
	for feat, val := range v.Vals {
		strs = append(strs, fmt.Sprintf("%v %v", feat, val))
	}
	return strings.Join(strs, "\n")
}

func (v *AvgSparse) Serialize() interface{} {
	// retval := make(map[interface{}][]int64, len(v.Vals))
	retval := make(map[interface{}][]int64, len(v.Vals))
	for k, v := range v.Vals {
		scores := make([]int64, len(v.Vals))
		for i, lastScore := range v.Vals {
			if lastScore != nil {
				scores[i] = lastScore.Value
			}
		}
		retval[k] = scores
	}
	return retval
}

func (v *AvgSparse) Deserialize(serialized interface{}, generation int) {
	data, ok := serialized.(map[interface{}][]int64)
	if !ok {
		panic("Can't deserialize unknown serialization")
	}
	v.Vals = make(map[Feature]*LockedArray, len(data))
	for k, datav := range data {
		lockedArray := &LockedArray{Vals: make([]*HistoryValue, len(datav))}
		for i, value := range datav {
			lockedArray.Vals[i] = NewHistoryValue(generation, value)
		}
		v.Vals[k] = lockedArray
	}
}

func NewAvgSparse() *AvgSparse {
	return &AvgSparse{Vals: make(map[Feature]*LockedArray, 100000)}
}
