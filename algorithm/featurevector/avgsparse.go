package featurevector

import (
	"fmt"
	// "log"
	"strings"
	"sync"
)

type HistoryValue struct {
	sync.Mutex
	Generation int
	Value      int64
	Previous   *HistoryValue
}

func (h *HistoryValue) Integrate(generation int) {
	if generation == 0 {
		panic("Cannot divide by generation 0")
	}
	if generation == h.Generation {
		if h.Previous != nil {
			h.Previous.Integrate(generation)
			h.Value = h.Previous.Value
			h.Previous = nil
		}
		return
	}
	var (
		curValue      int64       = 0.0 // explicitly initialize to 0.0
		curHistory    *HistoryValue = h
		curGeneration int           = generation
	)
	for curHistory != nil {
		curValue += curHistory.Value * int64(curGeneration-curHistory.Generation)
		curGeneration = curHistory.Generation
		curHistory = curHistory.Previous
	}
	h.Value = curValue // / int64(generation)
	h.Previous = nil
}

func (h *HistoryValue) Push(generation int) {
	newH := new(HistoryValue)
	*newH = *h
	h.Previous = newH
	h.Generation = generation
}

func (h *HistoryValue) Increment(generation int) {
	h.Add(generation, 1.0)
}

func (h *HistoryValue) Decrement(generation int) {
	h.Add(generation, -1.0)
}

func (h *HistoryValue) Add(generation int, amount int64) {
	h.Lock()
	defer h.Unlock()
	if generation > h.Generation {
		h.Push(generation)
	}
	h.Value = h.Value + amount
}

func NewHistoryValue(generation int, value int64) *HistoryValue {
	return &HistoryValue{Generation: generation, Value: value}
}

type LockedArray struct {
	sync.RWMutex
	vals []*HistoryValue
}

func (l *LockedArray) ExtendFor(generation, transition int) {
	newVals := make([]*HistoryValue, transition+1)
	copy(newVals[0:len(l.vals)], l.vals[0:len(l.vals)])
	l.vals = newVals
}

func (l *LockedArray) Add(generation, transition int, feature interface{}, amount int64) {
	l.Lock()
	defer l.Unlock()
	if transition < len(l.vals) {
		if l.vals[transition] != nil {
			l.vals[transition].Add(generation, amount)
		} else {
			l.vals[transition] = NewHistoryValue(generation, amount)
		}
		return
	} else {
		l.ExtendFor(generation, transition)
		if transition >= len(l.vals) {
			panic("Despite extending, transition >= than vals")
		}
		l.vals[transition] = NewHistoryValue(generation, amount)
		return
	}
}

type AvgSparse struct {
	sync.RWMutex
	vals map[Feature]*LockedArray
}

// func (v *AvgSparse) Copy() AvgSparse {
// 	copied := make(AvgSparse, len(v))
// 	for k, val := range v {
// 		copied[k] = val
// 	}
// 	return copied
// }

func (v *AvgSparse) Value(transition int, feature interface{}) int64 {
	transitions, exists := v.vals[feature]
	if exists && transition < len(transitions.vals) && transitions.vals[transition] != nil {
		return transitions.vals[transition].Value
	}
	return 0.0
}

func (v *AvgSparse) Add(generation, transition int, feature interface{}, amount int64, wg *sync.WaitGroup) {
	v.Lock()
	defer v.Unlock()
	transitions, exists := v.vals[feature]
	if exists {
		// wg.Add(1)
		go func() {
			transitions.Add(generation, transition, feature, amount)
			wg.Done()
		}()
	} else {
		newTrans := &LockedArray{vals: make([]*HistoryValue, transition+1)}
		newTrans.vals[transition] = NewHistoryValue(generation, amount)
		if v.vals == nil {
			panic("Got nil vals")
		}
		v.vals[feature] = newTrans
		wg.Done()
	}
}

func (v *AvgSparse) Integrate(generation int) *AvgSparse {
	for _, val := range v.vals {
		for _, transition := range val.vals {
			if transition != nil {
				transition.Integrate(generation)
			}
		}
	}
	return v
}

func (v *AvgSparse) SetScores(feature Feature, scores *[]int64) {
	transitions, exists := v.vals[feature]
	if exists {
		// log.Println("\t\tSetting scores for feature", feature)
		// log.Println("\t\t\t1. Exists")
		if cap(*scores) < len(transitions.vals) {
			// log.Println("\t\t\t1.1 Scores array not large enough")
			newscores := make([]int64, len(transitions.vals))
			// log.Println("\t\t\t1.2 Copying")
			copy(newscores[0:len(transitions.vals)], (*scores)[0:len(*scores)])
			// log.Println("\t\t\t1.3 Setting pointer")
			*scores = newscores
		}
		// log.Println("\t\t\t2. Iterating", len(transitions), "transitions")
		for i, val := range transitions.vals {
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
	for _, val := range v.vals {
		for _, transition := range val.vals {
			transition.Value = transition.Value / byValue
		}
	}
	return v
}

func (v *AvgSparse) String() string {
	strs := make([]string, 0, len(v.vals))
	v.RLock()
	defer v.RUnlock()
	for feat, val := range v.vals {
		strs = append(strs, fmt.Sprintf("%v %v", feat, val))
	}
	return strings.Join(strs, "\n")
}

func NewAvgSparse() *AvgSparse {
	return &AvgSparse{vals: make(map[Feature]*LockedArray, 100000)}
}
