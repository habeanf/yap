package Model

import (
	. "chukuparser/Algorithm/FeatureVector"
	"chukuparser/Algorithm/Perceptron"
	. "chukuparser/Algorithm/Transition"
	"fmt"
	// "sort"
	"strings"
)

type FeaturesList struct {
	Features   []Feature
	Transition Transition
	Previous   *FeaturesList
}

// func (f *FeaturesList) String() string {
// 	cur := f
// 	results := make([]string, 0, 10)
// 	for cur != nil && cur.Previous != nil {
// 		results = append(results, fmt.Sprintf("Trans,Size: %v, %v", cur.Transition, len(cur.Features)))
// 		cur = cur.Previous
// 	}
// 	// sort.Reverse(sort.StringSlice(results))
// 	return strings.Join(results, "\n")
// }

type Interface interface {
	Perceptron.Model
	TransitionScore(transition Transition, features []Feature) float64
}

func (l *FeaturesList) String() string {
	var (
		retval []string      = make([]string, 0, 100)
		cur    *FeaturesList = l
	)
	for cur != nil {
		retval = append(retval, fmt.Sprintf("%v", cur.Transition))
		for _, val := range cur.Features {
			retval = append(retval, fmt.Sprintf("\t%v", val))
		}
		cur = cur.Previous
	}
	return strings.Join(retval, "\n")
}

func MakeFeature(transition, i int, feat interface{}) interface{} {
	return [3]interface{}{transition, i, feat}
}
