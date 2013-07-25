package Model

type SparseWeightVector map[Feature]float64

func (v *SparseWeightVector) Add(other *SparseWeightVector) *SparseWeightVector {
	vec1 := *v
	retvec := vec1
	for key, otherVal := range *other {
		curVal, _ := vec1[key]
		retvec[key] = curVal + otherVal
	}
	// var wg sync.WaitGroup
	// wg.Add(len(*other))
	// for key, val := range *other {
	// 	go func(i int, val float64) {
	// 		defer wg.Done()
	// 		Bug, map is not threadsafe :(
	// 		vec[key] += val
	// 	}(key, val)
	// }
	// wg.Wait()
	return &retvec
}

func (v *SparseWeightVector) Subtract(other *SparseWeightVector) *SparseWeightVector {
	vec1 := *v
	retvec := vec1
	for key, otherVal := range *other {
		curVal, _ := vec1[key]
		retvec[key] = curVal - otherVal
	}
	// var wg sync.WaitGroup
	// wg.Add(len(*other))
	// for key, val := range *other {
	// 	go func(i int, val float64) {
	// 		defer wg.Done()
	// 		Bug, map is not threadsafe :(
	// 		vec[key] += val
	// 	}(key, val)
	// }
	// wg.Wait()
	return &retvec
}

func (v *SparseWeightVector) UpdateAdd(other *SparseWeightVector) *SparseWeightVector {
	vec := *v
	for key, otherVal := range *other {
		curVal, _ := vec[key]
		vec[key] = curVal + otherVal
	}
	// var wg sync.WaitGroup
	// wg.Add(len(*other))
	// for key, val := range *other {
	// 	go func(i int, val float64) {
	// 		defer wg.Done()
	// 		Bug, map is not threadsafe :(
	// 		vec[key] += val
	// 	}(key, val)
	// }
	// wg.Wait()
	return v
}

func (v *SparseWeightVector) UpdateSubtract(other *SparseWeightVector) *SparseWeightVector {
	vec := *v
	for key, otherVal := range *other {
		curVal, _ := vec[key]
		vec[key] = curVal - otherVal
	}
	// var wg sync.WaitGroup
	// wg.Add(len(*other))
	// for i, val := range *other {
	// 	go func(i int, val float64) {
	// 		defer wg.Done()
	// 		vec[i] -= val
	// 	}(i, val)
	// }
	// wg.Wait()
	return v
}

func (v *SparseWeightVector) DotProduct(other *SparseWeightVector) float64 {
	vec1 := *v
	vec2 := *other

	products := make(chan float64, len(vec2))
	for i, val := range vec2 {
		go func(result chan float64, val1, val2 float64) {
			result <- val1 * val2
		}(products, vec1[i], val)
	}
	close(products)

	var result float64
	for {
		val, ok := <-products
		if !ok {
			break
		}
		result += val
	}
	return result
}

func (v *SparseWeightVector) DotProductFeatures(f *[]Feature) float64 {
	vec1 := *v
	vec2 := *f
	products := make(chan float64, len(vec2))
	for _, val := range vec2 {
		go func(result chan float64, val Feature) {
			result <- vec1[val]
		}(products, val)
	}
	close(products)

	var result float64
	for {
		val, ok := <-products
		if !ok {
			break
		}
		result += val
	}
	return result
}

func (v *SparseWeightVector) FeatureWeights(f *[]Feature) *SparseWeightVector {
	vec1 := *v
	vec2 := *f
	retval := make(SparseWeightVector, len(vec2))
	for _, val := range vec2 {
		retval[val] = vec1[val]
	}
	return &retval
}
