package Perceptron

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
	for key, otherVal := range *other {
		// val[key] == 0 if val[key] does not exist
		retvec[key] = vec1[key] + otherVal
	}
	return &retvec
}

func (v *SparseWeightVector) Subtract(other *SparseWeightVector) *SparseWeightVector {
	vec1 := *v
	retvec := *(v.Copy())
	for key, otherVal := range *other {
		// val[key] == 0 if val[key] does not exist
		retvec[key] = vec1[key] - otherVal
	}
	return &retvec
}

func (v *SparseWeightVector) UpdateAdd(other *SparseWeightVector) *SparseWeightVector {
	vec := *v
	for key, otherVal := range *other {
		curVal, _ := vec[key]
		vec[key] = curVal + otherVal
	}
	return v
}

func (v *SparseWeightVector) UpdateSubtract(other *SparseWeightVector) *SparseWeightVector {
	vec := *v
	for key, otherVal := range *other {
		curVal, _ := vec[key]
		vec[key] = curVal - otherVal
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
