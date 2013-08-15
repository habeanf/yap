package Perceptron

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
)

type LinearPerceptron struct {
	Decoder    EarlyUpdateInstanceDecoder
	Updater    UpdateStrategy
	Iterations int
	Weights    *SparseWeightVector
	Log        bool
}

var _ SupervisedTrainer = &LinearPerceptron{}
var _ Model = &LinearPerceptron{}

func (m *LinearPerceptron) Score(features []Feature) float64 {
	return m.Weights.DotProductFeatures(features)
}

func (m *LinearPerceptron) Init() {
	// vec := make(SparseWeightVector, fe.EstimatedNumberOfFeatures())
	vec := make(SparseWeightVector)
	m.Weights = &vec
}

func (m *LinearPerceptron) Train(goldInstances []DecodedInstance) {
	m.train(goldInstances, m.Decoder, m.Iterations)
}

func (m *LinearPerceptron) train(goldInstances []DecodedInstance, decoder EarlyUpdateInstanceDecoder, iterations int) {
	if m.Weights == nil {
		panic("Model not initialized")
	}
	prevPrefix := log.Prefix()
	m.Updater.Init(m.Weights, iterations)
	for i := 0; i < iterations; i++ {
		log.SetPrefix("IT #" + fmt.Sprintf("%v ", i) + prevPrefix)
		if m.Log {
			log.Println("ITERATION", i)
		}
		for j, goldInstance := range goldInstances {
			if true {
				if m.Log && j%10 == 0 {
					log.Println("At instance", j)
				}
			}
			decodedInstance, decodedWeights, goldWeights := decoder.DecodeEarlyUpdate(goldInstance, m)
			if !goldInstance.Equal(decodedInstance) {
				if m.Log {
					// log.Println("Decoded did not equal gold, updating")
					// log.Println("Add Gold:")
					// for k, v := range *goldWeights {
					// 	log.Println(k, v)
					// }
					// log.Println()
					// log.Println("Sub Pred:")
					// for k, v := range *decodedWeights {
					// 	log.Println(k, v)
					// }
				}
				m.Weights.UpdateAdd(goldWeights).UpdateSubtract(decodedWeights)
				// log.Println()

				// log.Println("Weights after:")
				// for k, v := range *m.Weights {
				// 	log.Println(k, v)
				// }
				// log.Println()
			}
			m.Updater.Update(m.Weights)
		}
	}
	log.SetPrefix(prevPrefix)
	m.Weights = m.Updater.Finalize(m.Weights)
}

func (m *LinearPerceptron) Read(reader io.Reader) {
	dec := gob.NewDecoder(reader)
	model := make(SparseWeightVector)
	err := dec.Decode(&model)
	if err != nil {
		panic(err)
	}
	m.Weights = &model
}

func (m *LinearPerceptron) Write(writer io.Writer) {
	enc := gob.NewEncoder(writer)
	err := enc.Encode(m.Weights)
	if err != nil {
		panic(err)
	}
}

func (m *LinearPerceptron) String() string {
	return fmt.Sprintf("%v", *m.Weights)
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
	// explicitly reset u.N = 0.0 in case of reuse of vector
	// even though 0.0 is zero value
	u.N = 0.0
	u.P = float64(iterations)
	u.accumWeights = make(SparseWeightVector, len(*w))
}

func (u *AveragedStrategy) Update(w *SparseWeightVector) {
	u.accumWeights.UpdateAdd(w)
	u.N += 1
}

func (u *AveragedStrategy) Finalize(w *SparseWeightVector) *SparseWeightVector {
	u.accumWeights.UpdateScalarDivide(u.P * u.N)
	return &u.accumWeights
}
