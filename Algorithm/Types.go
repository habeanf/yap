package Algorithm

import "sync"

type Model interface {
	Score(i *DecodedInstance) float64
}

type Instance interface{}

type DecodedInstance interface {
	Instance
	Decode() interface{}
	SetInstance(i *Instance)
	SameDecode(other *DecodedInstance) bool
}

type Feature float64

type FeatureExtractor interface {
	Features(Instance) *[]Feature
	NumberOfFeatures() int
}

type Classifier interface {
	Classify(instance *Instance) *DecodedInstance
}

type Trainer interface {
	Train(instances chan *Instance)
}

type SupervisedTrainer interface {
	Train(instances chan *DecodedInstance)
}

type UnsupervisedTrainer interface {
	Train(instances chan *Instance)
}

type Decoder interface {
	Decode(i *Instance, m *Model) *DecodedInstance
}

type WeightVector []float64

func (v *WeightVector) Add(other *WeightVector) *WeightVector {
	var wg sync.WaitGroup
	wg.Add(len(*other))
	for i, val := range *other {
		go func(v *WeightVector, i int, val float64) {
			defer wg.Done()
			(*v)[i] = (*v)[i] + val
		}(v, i, val)
	}
	wg.Wait()
	return v
}

func (v *WeightVector) Subtract(other *WeightVector) *WeightVector {
	var wg sync.WaitGroup
	wg.Add(len(*other))
	for i, val := range *other {
		go func(v *WeightVector, i int, val float64) {
			defer wg.Done()
			(*v)[i] = (*v)[i] - val
		}(v, i, val)
	}
	wg.Wait()
	return v
}

func (v *WeightVector) DotProduct(other *WeightVector) float64 {
	temp := new(WeightVector)
	var wg sync.WaitGroup
	wg.Add(len(*other))
	for i, val := range *other {
		go func(t *WeightVector, v *WeightVector, i int, val float64) {
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

func (v *WeightVector) Init(val float64) {
	var wg sync.WaitGroup
	wg.Add(len(*v))
	for i, _ := range *v {
		go func(v *WeightVector, i int) {
			defer wg.Done()
			(*v)[i] = 0
		}(v, i)
	}
	wg.Wait()
}
