package Perceptron

type LinearPerceptron struct {
	Weights  SparseWeightVector
	FeatFunc FeatureExtractor
	Updater  UpdateStrategy
}

func (m *LinearPerceptron) Score(i *DecodedInstance) float64 {
	decodedFeatures := m.FeatFunc.Features((*i).GetInstance())
	return m.Weights.DotProductFeatures(decodedFeatures)
}

func (m *LinearPerceptron) Init(fe FeatureExtractor, up UpdateStrategy) {
	m.FeatFunc = fe
	m.Updater = up
	m.Weights = make(SparseWeightVector, fe.EstimatedNumberOfFeatures())
}

func (m *LinearPerceptron) Train(instances chan *DecodedInstance, decoder Decoder, iterations int) {
	if m.Weights == nil {
		panic("Model not initialized")
	}
	m.Updater.Init(&m.Weights, iterations)
	for instance := range instances {
		decodedInstance := decoder.Decode((*instance).GetInstance(), m)
		if !(*instance).Equals(decodedInstance) {
			decodedFeatures := m.FeatFunc.Features((*decodedInstance).GetInstance())
			goldFeatures := m.FeatFunc.Features((*instance).GetInstance())
			computedWeights := m.Weights.FeatureWeights(decodedFeatures)
			goldWeights := m.Weights.FeatureWeights(goldFeatures)
			m.Weights.UpdateAdd(goldWeights).UpdateSubtract(computedWeights)
		}
		m.Updater.Update(&m.Weights)
	}
	m.Weights = *m.Updater.Finalize(&m.Weights)
}

type UpdateStrategy interface {
	Init(w *SparseWeightVector, iterations int)
	Update(weights *SparseWeightVector)
	Finalize(w *SparseWeightVector) *SparseWeightVector
}

type TrivialStrategy struct{}

func (u *TrivialStrategy) Init(w *SparseWeightVector, iterations int) {

}

func (u *TrivialStrategy) Update(w *SparseWeightVector) {

}

func (u *TrivialStrategy) Finalize(w *SparseWeightVector) *SparseWeightVector {
	return w
}

type AveragedStrategy struct {
	P, N         float64
	accumWeights SparseWeightVector
}

func (u *AveragedStrategy) Init(w *SparseWeightVector, iterations int) {
	u.P = float64(iterations)
	u.N = 0
	u.accumWeights = make(SparseWeightVector, len(*w))
}

func (u *AveragedStrategy) Update(w *SparseWeightVector) {
	u.accumWeights.Add(w)
	u.N += 1
}

func (u *AveragedStrategy) Finalize(w *SparseWeightVector) *SparseWeightVector {
	for i, val := range u.accumWeights {
		u.accumWeights[i] = val / (u.P * u.N)
	}
	return &u.accumWeights
}
