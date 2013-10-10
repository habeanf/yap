package perceptron

import (
	// "chukuparser/algorithm/transition"
	"chukuparser/util"
	// "encoding/gob"
	"fmt"
	// "io"
	"log"
	// "os"
	"runtime"
)

type LinearPerceptron struct {
	Decoder        EarlyUpdateInstanceDecoder
	Updater        UpdateStrategy
	Iterations     int
	Model          Model
	Log            bool
	Tempfile       string
	TrainI, TrainJ int
	TempLines      int
}

var _ SupervisedTrainer = &LinearPerceptron{}

// var _ Model = &LinearPerceptron{}

// func (m *LinearPerceptron) Score(features []Feature) float64 {
// 	return m.Model.Score(features)
// }

func (m *LinearPerceptron) Init(newModel Model) {
	m.Model = newModel
	m.TrainI, m.TrainJ = 0, -1
	m.Updater.Init(m.Model, m.Iterations)
}

func (m *LinearPerceptron) Train(goldInstances []DecodedInstance) {
	m.train(goldInstances, m.Decoder, m.Iterations)
}

func (m *LinearPerceptron) train(goldInstances []DecodedInstance, decoder EarlyUpdateInstanceDecoder, iterations int) {
	if m.Model == nil {
		panic("Model not initialized")
	}
	allOut := false
	prevPrefix := log.Prefix()
	prevFlags := log.Flags()
	// var score float64
	for i := m.TrainI; i < iterations; i++ {
		log.SetPrefix("IT #" + fmt.Sprintf("%v ", i) + prevPrefix)
		if allOut {
			log.SetPrefix("")
			log.SetFlags(0)
		}
		for j, goldInstance := range goldInstances[m.TrainJ+1:] {
			// if m.Log {
			// 	if j%100 == 0 {
			// 		runtime.GC()
			// 	}
			// }
			decodedInstance, decodedFeatures, goldFeatures, earlyUpdatedAt, goldSize, score := decoder.DecodeEarlyUpdate(goldInstance, m.Model)
			if !goldInstance.Equal(decodedInstance) {
				if m.Log {
					// if allOut {
					// score = m.Model.Score(decodedFeatures)
					// }
					if earlyUpdatedAt >= 0 {
						if allOut {
							log.Printf("Error at %d of %d ; score %v\n", earlyUpdatedAt, goldSize, score)
						} else {
							log.Println("At instance", j, "failed", earlyUpdatedAt, "of", goldSize)
						}
					} else {
						if allOut {
							log.Printf("Error at %d of %d ; socre %v\n", goldSize, goldSize, score)
						} else {
							log.Println("At instance", j, "failed", goldSize, "of", goldSize)
						}
					}
					// log.Println("Decoded did not equal gold, updating")
					// log.Println("Decoded:")
					// log.Println(decodedInstance.Decoded())
					// log.Println("Gold:")
					// log.Println(goldInstance.Decoded())
					// if goldFeatures != nil {
					// 	log.Println("Add Gold:", goldFeatures, "features")
					// } else {
					// 	panic("Decode failed but got nil gold model")
					// }
					// if decodedFeatures != nil {
					// 	log.Println("Sub Pred:", decodedFeatures, "features")
					// } else {
					// 	panic("Decode failed but got nil decode model")
					// }
				}
				if allOut {
					log.Println("Score 1 to")
				}
				m.Model.AddSubtract(goldFeatures, decodedFeatures, 1.0)
				if allOut {
					log.Println("Score -1 to")
				}
				m.Model.AddSubtract(decodedFeatures, decodedFeatures, -1.0)
				// if m.Log {
				// 	log.Println("After Model Update:")
				// 	log.Println("\n", m.Model)
				// }
				// log.Println()

				// log.Println("Model after:")
				// for k, v := range *m.Model {
				// 	log.Println(k, v)
				// }
				// log.Println()
			} else {
				if m.Log {
					log.Println("At instance", j, "success")
				}
			}
			m.Updater.Update(m.Model)
			if m.TempLines > 0 && j > 0 && j%m.TempLines == 0 {
				// m.TrainJ = j
				// m.TrainI = i
				// if m.Log {
				// 	log.Println("Dumping at iteration", i, "after sent", j)
				// }
				// m.TempDump(m.Tempfile)
				if m.Log {
					log.Println("\tBefore GC")
					util.LogMemory()
					log.Println("\tRunning GC")
				}
				runtime.GC()
				if m.Log {
					log.Println("\tAfter GC")
					util.LogMemory()
					log.Println("\tDone GC")
				}
			}
		}

		// if m.Log {
		// 	log.Println("\tBefore GC")
		// 	util.LogMemory()
		// 	log.Println("\tRunning GC")
		// }
		if allOut {
			log.Println("ITERATION COMPLETE")
		}
		runtime.GC()
		// if m.Log {
		// 	log.Println("\tAfter GC")
		// 	util.LogMemory()
		// 	log.Println("\tDone GC")
		// }
	}
	log.SetPrefix(prevPrefix)
	log.SetFlags(prevFlags)
	m.Model = m.Updater.Finalize(m.Model)
}

// func (m *LinearPerceptron) Read(reader io.Reader) {
// 	dec := gob.NewDecoder(reader)
// 	model := make(Model)
// 	err := dec.Decode(&model)
// 	if err != nil {
// 		panic(err)
// 	}
// 	m.Model = &model
// }

// func (m *LinearPerceptron) TempDump(filename string) {
// 	log.Println("Temp dumping to", filename)
// 	file, err := os.Create(filename)
// 	defer file.Close()
// 	if err != nil {
// 		panic("Can't open file for temp write: " + err.Error())
// 	}
// 	enc := gob.NewEncoder(file)
// 	gobM := &LinearPerceptron{
// 		Updater:    m.Updater,
// 		TrainI:     m.TrainI,
// 		TrainJ:     m.TrainJ,
// 		TempLines:  m.TempLines,
// 		Tempfile:   m.Tempfile,
// 		Log:        m.Log,
// 		Iterations: m.Iterations,
// 		Weights:    m.Weights,
// 	}
// 	encErr := enc.Encode(gobM)
// 	if encErr != nil {
// 		panic("Failed to encode self: " + encErr.Error())
// 	}
// }

// func (m *LinearPerceptron) TempLoad(filename string) {
// 	log.Println("Temp loading from", filename)
// 	file, err := os.Open(filename)
// 	defer file.Close()
// 	if err != nil {
// 		panic("Can't open file for temp read: " + err.Error())
// 	}
// 	dec := gob.NewDecoder(file)
// 	decErr := dec.Decode(m)
// 	if decErr != nil {
// 		panic("Failed to decode self: " + decErr.Error())
// 	}
// 	log.Println("Done")
// 	log.Println("Iteration #, Train Instance:", m.TrainI, m.TrainJ)
// }

// func (m *LinearPerceptron) Write(writer io.Writer) {
// 	enc := gob.NewEncoder(writer)
// 	err := enc.Encode(m.Weights)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func (m *LinearPerceptron) String() string {
// 	return fmt.Sprintf("%v", m.Model)
// }

type UpdateStrategy interface {
	Init(m Model, iterations int)
	Update(model Model)
	Finalize(m Model) Model
}

type TrivialStrategy struct{}

func (u *TrivialStrategy) Init(m Model, iterations int) {

}

func (u *TrivialStrategy) Update(m Model) {

}

func (u *TrivialStrategy) Finalize(m Model) Model {
	return m
}

type AveragedStrategy struct {
	P, N       float64
	accumModel Model
}

func (u *AveragedStrategy) Init(m Model, iterations int) {
	// explicitly reset u.N = 0.0 in case of reuse of vector
	// even though 0.0 is zero value
	u.N = 0.0
	u.P = float64(iterations)
	u.accumModel = m.New()
}

func (u *AveragedStrategy) Update(m Model) {
	u.accumModel.AddModel(m)
	u.N += 1
}

func (u *AveragedStrategy) Finalize(m Model) Model {
	u.accumModel.ScalarDivide(u.P * u.N)
	return u.accumModel
}
