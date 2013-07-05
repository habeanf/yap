package Algorithm

type LinearPerceptron struct {
	Weights  *SparseWeightVector
	FeatFunc *FeatureExtractor
	Updater  *UpdateStrategy
}

func (m *LinearPerceptron) Score(i *DecodedInstance) float64 {
	decodedFeatures := m.FeatFunc.Features(i)
	return m.Weights.DotProduct(decodedFeatures)
}

func (m *LinearPerceptron) Init(fe *FeatureExtractor, up *UpdateStrategy) {
	m.FeatFunc = fe
	m.Updater = up
	m.Weights = make(SparseWeightVector, fe.NumberOfFeatures())
	m.Weights.Init(0)
}

func (m *LinearPerceptron) Train(instances chan *DecodedInstance, decoder Decoder, iterations int) {
	if m.NumFeatures == 0 {
		panic("Model not initialized")
	}
	var result bool
	m.Updater.Init(m.Weights)
	for instance := range instances {
		decodedInstance := decoder.Decode(instance, m.Weights)
		if !instance.Equals(decoded) {
			computedWeights = m.Score(decodedInstance)
			trueWeights = m.Score(actualInstance)
			m.Weights = m.Weights.Add(trueWeights).Subtract(computedWeights)
		}
		m.Updater.Update(m.Weights)
	}
	return m.Updater.Finalize(m.Weights)
}

type UpdateStrategy interface {
	Init(w *SparseWeightVector, iterations int, instances int)
	Update(weights *SparseWeightVector)
	Finalize(w *SparseWeightVector) *SparseWeightVector
}

type TrivialStrategy struct{}

func (u *TrivialStrategy) Init(w *SparseWeightVector, iterations int, instances int) {

}

func (u *TrivialStrategy) Update(w *SparseWeightVector) {

}

func (u *TrivialStrategy) Finalize(w *SparseWeightVector) *SparseWeightVector {
	return w
}

type AveragedStrategy struct {
	P, N         float64
	accumWeights *SparseWeightVector
}

func (u *AveragedStrategy) Init(w *SparseWeightVector, iterations int, instances int) {
	u.P = float64(iterations)
	u.N = float64(instances)
	u.accumWeights = make(SparseWeightVector, len(*w))
}

func (u *AveragedStrategy) Update(w *SparseWeightVector) {
	u.accumWeights.Add(w)
}

func (u *AveragedStrategy) Finalize(w *SparseWeightVector) *SparseWeightVector {
	for i, val := range u.accumWeights {
		u.accumWeights[i] = val / (u.P * u.N)
	}
	return u.accumWeights
}
