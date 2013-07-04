package Algorithm

type SparseWeightVector map[string]float64

func (v *SparseWeightVector) Add(other *SparseWeightVector) *SparseWeightVector {
	var wg sync.WaitGroup
	wg.Add(len(*other))
	for i, val := range *other {
		go func(v *SparseWeightVector, i int, val float64) {
			defer wg.Done()
			(*v)[i] = (*v)[i] + val
		}(v, i, val)
	}
	wg.Wait()
	return v
}

func (v *SparseWeightVector) Subtract(other *SparseWeightVector) *SparseWeightVector {
	var wg sync.WaitGroup
	wg.Add(len(*other))
	for i, val := range *other {
		go func(v *SparseWeightVector, i int, val float64) {
			defer wg.Done()
			(*v)[i] = (*v)[i] - val
		}(v, i, val)
	}
	wg.Wait()
	return v
}

func (v *SparseWeightVector) DotProduct(other *SparseWeightVector) float64 {
	temp := new(SparseWeightVector)
	var wg sync.WaitGroup
	wg.Add(len(*other))
	for i, val := range *other {
		go func(t *SparseWeightVector, v *SparseWeightVector, i int, val float64) {
			defer wg.Done()
			(*t)[i] = (*v)[i] * val
		}(t, v, i, val)
	}
	wg.Wait()

	var result float64 = 0
	for _, val := range *temp {
		result += val
	}
	return result
}

func (v *SparseWeightVector) Init(val float64) {
	var wg sync.WaitGroup
	wg.Add(len(*v))
	for i, _ := range *v {
		go func(v *SparseWeightVector, i int) {
			defer wg.Done()
			(*v)[i] = 0
		}(v, i)
	}
	wg.Wait()
}
