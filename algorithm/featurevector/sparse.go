package featurevector

import (
	"fmt"
	// "log"
	"strings"
	// "sync"
)

type Sparse map[Feature]int64

func (v Sparse) Copy() Sparse {
	copied := make(Sparse, len(v))
	for k, val := range v {
		copied[k] = val
	}
	return copied
}

func (v Sparse) Add(other Sparse) Sparse {
	vec1 := v
	retvec := (v.Copy())
	var val int64
	if other == nil {
		return retvec
	}
	for key, otherVal := range other {
		// val[key] == 0 if val[key] does not exist
		val = vec1[key] + otherVal
		if val != 0.0 {
			retvec[key] = val
		} else {
			delete(retvec, key)
		}
	}
	return retvec
}

func (v Sparse) Subtract(other Sparse) Sparse {
	vec1 := v
	retvec := (v.Copy())
	var val int64
	if other == nil {
		return retvec
	}
	for key, otherVal := range other {
		// vec1[key] == 0 if vec1[key] does not exist
		val = vec1[key] - otherVal
		if val != 0.0 {
			retvec[key] = val
		} else {
			delete(retvec, key)
		}
	}
	return retvec
}

func (v Sparse) UpdateAdd(other Sparse) Sparse {
	vec := v
	if other == nil {
		return v
	}
	var val int64

	for key, otherVal := range other {
		val = vec[key] + otherVal
		if val != 0.0 {
			vec[key] = val
		} else {
			delete(vec, key)
		}
	}
	return v
}

func (v Sparse) UpdateSubtract(other Sparse) Sparse {
	vec := v
	if other == nil {
		return v
	}
	var val int64

	for key, otherVal := range other {
		val = vec[key] - otherVal
		if val != 0.0 {
			vec[key] = val
		} else {
			delete(vec, key)
		}
	}
	return v
}

func (v Sparse) UpdateScalarDivide(byValue int64) Sparse {
	if byValue == 0.0 {
		panic("Divide by 0")
	}
	vec := v
	for i, val := range vec {
		vec[i] = val / byValue
	}
	return v
}

func (v Sparse) DotProduct(other Sparse) int64 {
	vec1 := v
	vec2 := other

	var result int64
	for i, val := range vec2 {
		// val[i] == 0 if val[i] does not exist
		result += vec1[i] * val
	}
	return result
}

func (v Sparse) DotProductFeatures(f []Feature) int64 {
	vec1 := v
	vec2 := f

	var result int64
	for _, val := range vec2 {
		result += vec1[val]
	}
	return result
}

func (v Sparse) Weighted(other Sparse) Sparse {
	vec1 := v
	retvec := make(Sparse, len(other))
	if other == nil {
		return retvec
	}
	for key, otherVal := range other {
		// val[key] == 0 if val[key] does not exist
		retvec[key] = vec1[key] * otherVal
	}
	return retvec

}

func (v Sparse) FeatureWeights(f []Feature) Sparse {
	vec1 := v
	vec2 := f
	retval := make(Sparse, len(vec2))
	for _, val := range vec2 {
		retval[val] = vec1[val]
	}
	return retval
}

func (v Sparse) L1Norm() int64 {
	vec1 := v

	var result int64
	for _, val := range vec1 {
		result += val
	}
	return result
}

func (v Sparse) String() string {
	strs := make([]string, 0, len(v))
	for feat, val := range v {
		strs = append(strs, fmt.Sprintf("%v %v", feat, val))
	}
	return strings.Join(strs, "\n")
}

func NewVectorOfOnesFromFeatures(f []Feature) Sparse {
	vec := make(Sparse, len(f))
	for _, feature := range f {
		vec[feature] = 1.0
	}
	return vec
}

func NewSparse() Sparse {
	return make(Sparse)
}
