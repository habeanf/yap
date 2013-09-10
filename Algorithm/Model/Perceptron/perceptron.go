package Perceptron

import (
	"chukuparser/Util"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
)

type LinearPerceptron struct {
	Decoder        EarlyUpdateInstanceDecoder
	Updater        UpdateStrategy
	Iterations     int
	Weights        *SparseWeightVector
	Log            bool
	Tempfile       string
	TrainI, TrainJ int
	TempLines      int
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
	m.TrainI, m.TrainJ = 0, -1
	m.Updater.Init(m.Weights, m.Iterations)
}

func (m *LinearPerceptron) Train(goldInstances []DecodedInstance) {
	m.train(goldInstances, m.Decoder, m.Iterations)
}

func (m *LinearPerceptron) train(goldInstances []DecodedInstance, decoder EarlyUpdateInstanceDecoder, iterations int) {
	if m.Weights == nil {
		panic("Model not initialized")
	}
	prevPrefix := log.Prefix()
	for i := m.TrainI; i < iterations; i++ {
		log.SetPrefix("IT #" + fmt.Sprintf("%v ", i) + prevPrefix)
		for j, goldInstance := range goldInstances[m.TrainJ+1:] {
			if m.Log {
				if j%100 == 0 {
					log.Println("At instance", j)
					runtime.GC()
				}
			}
			decodedInstance, decodedWeights, goldWeights := decoder.DecodeEarlyUpdate(goldInstance, m)
			if !goldInstance.Equal(decodedInstance) {
				if m.Log {
					// log.Println("Decoded did not equal gold, updating")
					// log.Println("Decoded:")
					// log.Println(decodedInstance.Instance())
					// log.Println("Gold:")
					// log.Println(goldInstance.Instance())
					// if goldWeights != nil {
					// 	log.Println("Add Gold:", len(*goldWeights), "features")
					// } else {
					// 	panic("Decode failed but got nil gold weights")
					// }
					// if decodedWeights != nil {
					// 	log.Println("Sub Pred:", len(*decodedWeights), "features")
					// } else {
					// 	panic("Decode failed but got nil decode weights")
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
			if m.TempLines > 0 && j > 0 && j%m.TempLines == 0 {
				// m.TrainJ = j
				// m.TrainI = i
				// if m.Log {
				// 	log.Println("Dumping at iteration", i, "after sent", j)
				// }
				// m.TempDump(m.Tempfile)
				if m.Log {
					log.Println("\tBefore GC")
					Util.LogMemory()
					log.Println("\tRunning GC")
				}
				runtime.GC()
				if m.Log {
					log.Println("\tAfter GC")
					Util.LogMemory()
					log.Println("\tDone GC")
				}
			}
		}

		// if m.Log {
		// 	log.Println("\tBefore GC")
		// 	Util.LogMemory()
		// 	log.Println("\tRunning GC")
		// }
		runtime.GC()
		// if m.Log {
		// 	log.Println("\tAfter GC")
		// 	Util.LogMemory()
		// 	log.Println("\tDone GC")
		// }
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

func (m *LinearPerceptron) TempDump(filename string) {
	log.Println("Temp dumping to", filename)
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		panic("Can't open file for temp write: " + err.Error())
	}
	enc := gob.NewEncoder(file)
	gobM := &LinearPerceptron{
		Updater:    m.Updater,
		TrainI:     m.TrainI,
		TrainJ:     m.TrainJ,
		TempLines:  m.TempLines,
		Tempfile:   m.Tempfile,
		Log:        m.Log,
		Iterations: m.Iterations,
		Weights:    m.Weights,
	}
	encErr := enc.Encode(gobM)
	if encErr != nil {
		panic("Failed to encode self: " + encErr.Error())
	}
}

func (m *LinearPerceptron) TempLoad(filename string) {
	log.Println("Temp loading from", filename)
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		panic("Can't open file for temp read: " + err.Error())
	}
	dec := gob.NewDecoder(file)
	decErr := dec.Decode(m)
	if decErr != nil {
		panic("Failed to decode self: " + decErr.Error())
	}
	log.Println("Done")
	log.Println("Iteration #, Train Instance:", m.TrainI, m.TrainJ)
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
