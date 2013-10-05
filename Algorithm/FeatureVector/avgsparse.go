package FeatureVector

import (
	"fmt"
	// "log"
	"strings"
	// "sync"
)

type HistoryValue struct {
	Generation int
	Value      float64
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
		curValue      float64       = 0.0 // explicitly initialize to 0.0
		curHistory    *HistoryValue = h
		curGeneration int           = generation
	)
	for curHistory != nil {
		curValue += curHistory.Value * float64(curGeneration-curHistory.Generation)
		curGeneration = curHistory.Generation
		curHistory = curHistory.Previous
	}
	h.Value = curValue / float64(generation)
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

func (h *HistoryValue) Add(generation int, amount float64) {
	if generation > h.Generation {
		h.Push(generation)
	}
	h.Value = h.Value + amount
}

func NewHistoryValue(generation int, value float64) *HistoryValue {
	return &HistoryValue{generation, value, nil}
}

type AvgSparse map[Feature][]*HistoryValue

// func (v AvgSparse) Copy() AvgSparse {
// 	copied := make(AvgSparse, len(v))
// 	for k, val := range v {
// 		copied[k] = val
// 	}
// 	return copied
// }

func (v AvgSparse) Value(transition int, feature interface{}) float64 {
	transitions, exists := v[feature]
	if exists && transition < len(transitions) && transitions[transition] != nil {
		return transitions[transition].Value
	}
	return 0.0
}

func (v AvgSparse) Add(generation, transition int, feature interface{}, amount float64) {
	transitions, exists := v[feature]
	if exists {
		if transition < len(transitions) {
			if transitions[transition] != nil {
				transitions[transition].Add(generation, amount)
			} else {
				transitions[transition] = NewHistoryValue(generation, amount)
			}
			return
		} else {
			newTrans := make([]*HistoryValue, transition+1)
			copy(newTrans[0:len(transitions)], transitions[0:len(transitions)])
			v[feature] = newTrans
		}
	} else {
		v[feature] = make([]*HistoryValue, transition+1)
	}
	v[feature][transition] = NewHistoryValue(generation, amount)
}

func (v AvgSparse) Integrate(generation int) AvgSparse {
	for _, val := range v {
		for _, transition := range val {
			if transition != nil {
				transition.Integrate(generation)
			}
		}
	}
	return v
}

func (v AvgSparse) SetScores(feature Feature, scores *[]float64) {
	transitions, exists := v[feature]
	if exists {
		// log.Println("\t\tSetting scores for feature", feature)
		// log.Println("\t\t\t1. Exists")
		if cap(*scores) < len(transitions) {
			// log.Println("\t\t\t1.1 Scores array not large enough")
			newscores := make([]float64, len(transitions))
			// log.Println("\t\t\t1.2 Copying")
			copy(newscores[0:len(transitions)], (*scores)[0:len(*scores)])
			// log.Println("\t\t\t1.3 Setting pointer")
			*scores = newscores
		}
		// log.Println("\t\t\t2. Iterating", len(transitions), "transitions")
		for i, val := range transitions {
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

func (v AvgSparse) UpdateScalarDivide(byValue float64) AvgSparse {
	if byValue == 0.0 {
		panic("Divide by 0")
	}
	for _, val := range v {
		for _, transition := range val {
			transition.Value = transition.Value / byValue
		}
	}
	return v
}

func (v AvgSparse) String() string {
	strs := make([]string, 0, len(v))
	for feat, val := range v {
		strs = append(strs, fmt.Sprintf("%v %v", feat, val))
	}
	return strings.Join(strs, "\n")
}

func NewAvgSparse() AvgSparse {
	return make(AvgSparse)
}
