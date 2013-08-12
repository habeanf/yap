package Perceptron

import (
	"fmt"
	"strings"
	// "sync"
)

type SparseWeightVector map[Feature]float64

func (v *SparseWeightVector) Copy() *SparseWeightVector {
	copied := make(SparseWeightVector, len(*v))
	for k, val := range *v {
		copied[k] = val
	}
	return &copied
}

func (v *SparseWeightVector) Add(other *SparseWeightVector) *SparseWeightVector {
	vec1 := *v
	retvec := *(v.Copy())
	if other == nil {
		return &retvec
	}
	for key, otherVal := range *other {
		// val[key] == 0 if val[key] does not exist
		retvec[key] = vec1[key] + otherVal
	}
	return &retvec
}

func (v *SparseWeightVector) Subtract(other *SparseWeightVector) *SparseWeightVector {
	vec1 := *v
	retvec := *(v.Copy())
	if other == nil {
		return &retvec
	}
	for key, otherVal := range *other {
		// vec1[key] == 0 if vec1[key] does not exist
		retvec[key] = vec1[key] - otherVal
	}
	return &retvec
}

func (v *SparseWeightVector) UpdateAdd(other *SparseWeightVector) *SparseWeightVector {
	vec := *v
	if other == nil {
		return v
	}
	for key, otherVal := range *other {
		vec[key] = vec[key] + otherVal
	}
	return v
}

func (v *SparseWeightVector) UpdateSubtract(other *SparseWeightVector) *SparseWeightVector {
	vec := *v
	if other == nil {
		return v
	}
	for key, otherVal := range *other {
		vec[key] = vec[key] - otherVal
	}
	return v
}

func (v *SparseWeightVector) UpdateScalarDivide(byValue float64) *SparseWeightVector {
	if byValue == 0.0 {
		panic("Divide by 0")
	}
	vec := *v
	for i, val := range vec {
		vec[i] = val / byValue
	}
	return v
}

func (v *SparseWeightVector) DotProduct(other *SparseWeightVector) float64 {
	vec1 := *v
	vec2 := *other

	var result float64
	for i, val := range vec2 {
		// val[i] == 0 if val[i] does not exist
		result += vec1[i] * val
	}
	return result
}

func (v *SparseWeightVector) DotProductFeatures(f []Feature) float64 {
	vec1 := *v
	vec2 := f

	var result float64
	for _, val := range vec2 {
		result += vec1[val]
	}
	return result
}

func (v *SparseWeightVector) FeatureWeights(f []Feature) *SparseWeightVector {
	vec1 := *v
	vec2 := f
	retval := make(SparseWeightVector, len(vec2))
	for _, val := range vec2 {
		retval[val] = vec1[val]
	}
	return &retval
}

func (v *SparseWeightVector) L1Norm() float64 {
	vec1 := *v

	var result float64
	for _, val := range vec1 {
		result += val
	}
	return result
}

func (v *SparseWeightVector) String() string {
	strs := make([]string, 0, len(*v))
	for feat, val := range *v {
		strs = append(strs, fmt.Sprintf("%v %v", feat, val))
	}
	return strings.Join(strs, "\n")
}

func NewVectorOfOnesFromFeatures(f []Feature) *SparseWeightVector {
	vec := make(SparseWeightVector, len(f))
	for _, feature := range f {
		vec[feature] = 1.0
	}
	return &vec
}
