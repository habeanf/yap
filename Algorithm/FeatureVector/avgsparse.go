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
	var (
		curValue      float64       = 0.0 // explicitly initialize to 0.0
		curHistory    *HistoryValue = h
		curGeneration int           = generation
	)
	for curHistory != nil {
		if curHistory.Previous != nil {
			curValue += curHistory.Value * (curGeneration - curHistory.Generation)
		}
		curHistory = curHistory.Previous
		curGeneration = curHistory.Generation
	}
	h.Value = curValue / float64(generation)
}

func (h *HistoryValue) Push() {
	newH := new(HistoryValue)
	*newH = *h
	h.Previous = newH
}

func (h *HistoryValue) Increment(generation int) {
	if generation > h.Generation {
		h.Push()
	}
	h.Value = h.Value + 1.0
}

func (h *HistoryValue) Decrement(generation int) {
	if generation > h.Generation {
		h.Push()
	}
	h.Value = h.Value - 1.0
}

func NewHistoryValue(generation, value int) {
	return &HistoryValue{generation, value, nil}
}

type AvgSparse map[Feature]*HistoryValue

func (v AvgSparse) Copy() AvgSparse {
	copied := make(AvgSparse, len(v))
	for k, val := range v {
		copied[k] = val
	}
	return copied
}

func (v AvgSparse) Value(feature interface{}) float64 {
	val, exists := v[feature]
	if exists {
		return val.Value
	} else {
		return 0.0
	}
}

func (v AvgSparse) Increment(generation int, feature interface{}) {
	val, exists := v[feature]
	if exists {
		val.Increment(generation)
	} else {
		v[feature] = NewHistoryValue(generation, 1.0)
	}
}

func (v AvgSparse) Decrement(generation int, feature interface{}) {
	val, exists := v[feature]
	if exists {
		val.Decrement(generation)
	} else {
		v[feature] = NewHistoryValue(generation, -1.0)
	}

}

func (v AvgSparse) Integrate(generation int) {
	for _, val := range v {
		Integrate(generation)
	}
	return v
}

func (v AvgSparse) UpdateScalarDivide(byValue float64) AvgSparse {
	if byValue == 0.0 {
		panic("Divide by 0")
	}
	for _, val := range v {
		val.Value = val.Value / byValue
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
