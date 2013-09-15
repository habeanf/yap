package Perceptron

import (
	"fmt"
	// "log"
	"strings"
	// "sync"
)

type SparseFeatureVector map[Feature]float64

func (v *SparseFeatureVector) Copy() *SparseFeatureVector {
	copied := make(SparseFeatureVector, len(*v))
	for k, val := range *v {
		copied[k] = val
	}
	return &copied
}

func (v *SparseFeatureVector) Add(other *SparseFeatureVector) *SparseFeatureVector {
	vec1 := *v
	retvec := *(v.Copy())
	var val float64
	if other == nil {
		return &retvec
	}
	for key, otherVal := range *other {
		// val[key] == 0 if val[key] does not exist
		val = vec1[key] + otherVal
		if val != 0.0 {
			retvec[key] = val
		} else {
			delete(retvec, key)
		}
	}
	return &retvec
}

func (v *SparseFeatureVector) Subtract(other *SparseFeatureVector) *SparseFeatureVector {
	vec1 := *v
	retvec := *(v.Copy())
	var val float64
	if other == nil {
		return &retvec
	}
	for key, otherVal := range *other {
		// vec1[key] == 0 if vec1[key] does not exist
		val = vec1[key] - otherVal
		if val != 0.0 {
			retvec[key] = val
		} else {
			delete(retvec, key)
		}
	}
	return &retvec
}

func (v *SparseFeatureVector) UpdateAdd(other *SparseFeatureVector) *SparseFeatureVector {
	vec := *v
	if other == nil {
		return v
	}
	var val float64

	for key, otherVal := range *other {
		val = vec[key] + otherVal
		if val != 0.0 {
			vec[key] = val
		} else {
			delete(vec, key)
		}
	}
	return v
}

func (v *SparseFeatureVector) UpdateSubtract(other *SparseFeatureVector) *SparseFeatureVector {
	vec := *v
	if other == nil {
		return v
	}
	var val float64

	for key, otherVal := range *other {
		val = vec[key] - otherVal
		if val != 0.0 {
			vec[key] = val
		} else {
			delete(vec, key)
		}
	}
	return v
}

func (v *SparseFeatureVector) UpdateScalarDivide(byValue float64) *SparseFeatureVector {
	if byValue == 0.0 {
		panic("Divide by 0")
	}
	vec := *v
	for i, val := range vec {
		vec[i] = val / byValue
	}
	return v
}

func (v *SparseFeatureVector) DotProduct(other *SparseFeatureVector) float64 {
	vec1 := *v
	vec2 := *other

	var result float64
	for i, val := range vec2 {
		// val[i] == 0 if val[i] does not exist
		result += vec1[i] * val
	}
	return result
}

func (v *SparseFeatureVector) DotProductFeatures(f []Feature) float64 {
	vec1 := *v
	vec2 := f

	var result float64
	for _, val := range vec2 {
		result += vec1[val]
	}
	return result
}

func (v *SparseFeatureVector) Weighted(other *SparseFeatureVector) *SparseFeatureVector {
	vec1 := *v
	retvec := make(SparseFeatureVector, len(*other))
	if other == nil {
		return &retvec
	}
	for key, otherVal := range *other {
		// val[key] == 0 if val[key] does not exist
		retvec[key] = vec1[key] * otherVal
	}
	return &retvec

}

func (v *SparseFeatureVector) FeatureWeights(f []Feature) *SparseFeatureVector {
	vec1 := *v
	vec2 := f
	retval := make(SparseFeatureVector, len(vec2))
	for _, val := range vec2 {
		retval[val] = vec1[val]
	}
	return &retval
}

func (v *SparseFeatureVector) L1Norm() float64 {
	vec1 := *v

	var result float64
	for _, val := range vec1 {
		result += val
	}
	return result
}

func (v *SparseFeatureVector) String() string {
	strs := make([]string, 0, len(*v))
	for feat, val := range *v {
		strs = append(strs, fmt.Sprintf("%v %v", feat, val))
	}
	return strings.Join(strs, "\n")
}

func NewVectorOfOnesFromFeatures(f []Feature) *SparseFeatureVector {
	vec := make(SparseFeatureVector, len(f))
	for _, feature := range f {
		vec[feature] = 1.0
	}
	return &vec
}
