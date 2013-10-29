package model

import (
	. "chukuparser/algorithm/featurevector"
	"chukuparser/algorithm/perceptron"
	"chukuparser/algorithm/transition"
	"chukuparser/util"
	"log"
	"sync"
)

var AllOut bool = false

type AvgMatrixSparse struct {
	Mat                  []*AvgSparse
	Features, Generation int
	Formatters           []util.Format
	Log                  bool
}

var _ perceptron.Model = &AvgMatrixSparse{}
var _ Interface = &AvgMatrixSparse{}

func (t *AvgMatrixSparse) Score(features interface{}) int64 {
	var (
		retval    int64
		intTrans  int
		prevScore int64
	)
	f := features.(*transition.FeaturesList)
	if f.Previous == nil {
		return 0
	}
	prevScore = t.Score(f.Previous)
	lastTransition := f.Transition
	featuresList := f.Previous
	intTrans = int(lastTransition)
	for i, feature := range featuresList.Features {
		if feature != nil {
			retval += t.Mat[i].Value(intTrans, feature)
		}
	}
	return prevScore + retval
}

func (t *AvgMatrixSparse) Add(features interface{}) perceptron.Model {
	if t.Log {
		log.Println("Score", 1.0, "to")
	}
	t.apply(features, 1.0)
	return t
}

func (t *AvgMatrixSparse) Subtract(features interface{}) perceptron.Model {
	if t.Log {
		log.Println("Score", -1.0, "to")
	}
	t.apply(features, -1.0)
	return t
}

func (t *AvgMatrixSparse) AddSubtract(goldFeatures, decodedFeatures interface{}, amount int64) {
	g := goldFeatures.(*transition.FeaturesList)
	f := decodedFeatures.(*transition.FeaturesList)
	if f.Previous == nil {
		return
	}
	// TODO: fix this hack
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		t.AddSubtract(g.Previous, f.Previous, amount)
		if t.Log {
			log.Println("\tstate", g.Transition)
		}
		wg.Done()
	}()
	if t.Log {
		wg.Wait()
	}
	t.apply(goldFeatures, amount)

	wg.Wait()
}

func (t *AvgMatrixSparse) apply(features interface{}, amount int64) perceptron.Model {
	var (
		intTrans int
	)
	f := features.(*transition.FeaturesList)
	if f.Previous == nil {
		return t
	}
	lastTransition := f.Transition
	featuresList := f.Previous
	// for featuresList != nil {
	intTrans = int(lastTransition)
	if intTrans >= 96 {
		return t
	}
	var wg sync.WaitGroup
	for i, feature := range featuresList.Features {
		if feature != nil {
			if t.Log {
				featTemp := t.Formatters[i]
				if t.Formatters != nil {
					log.Printf("\t\t%s %v %v\n", featTemp, featTemp.Format(feature), amount)
				}
			}
			wg.Add(1)
			go func(j int, feat interface{}) {
				t.Mat[j].Add(t.Generation, intTrans, feat, amount, &wg)
				// t.Mat[i].Add(t.Generation, intTrans, feature, amount, &wg)
				// wg.Done()
			}(i, feature)
			if AllOut {
				wg.Wait()
			}
		}
	}
	wg.Wait()
	// 	lastTransition = featuresList.Transition
	// 	featuresList = featuresList.Previous
	// }
	return t
}

func (t *AvgMatrixSparse) ScalarDivide(val int64) {
	for _, avgsparse := range t.Mat {
		avgsparse.UpdateScalarDivide(val)
	}
}

func (t *AvgMatrixSparse) Integrate() {
	for _, val := range t.Mat {
		val.Integrate(t.Generation)
	}
}

func (t *AvgMatrixSparse) IncrementGeneration() {
	t.Generation += 1
}

func (t *AvgMatrixSparse) Copy() perceptron.Model {
	panic("Cannot copy an avg matrix sparse representation")
	return nil
}

func (t *AvgMatrixSparse) New() perceptron.Model {
	return NewAvgMatrixSparse(t.Features, nil)
}

func (t *AvgMatrixSparse) AddModel(m perceptron.Model) {
	panic("Cannot add two avg matrix sparse types")
}

func (t *AvgMatrixSparse) TransitionScore(transition transition.Transition, features []Feature) int64 {
	var (
		retval   int64
		intTrans int = int(transition)
	)

	if len(features) > len(t.Mat) {
		panic("Got more features than known matrix features")
	}
	for i, feat := range features {
		if feat != nil {
			// val := t.Mat[i].Value(intTrans, feat)
			// if t.Formatters != nil {
			// 	featTemp := t.Formatters[i]
			// 	log.Printf("\t\t\t%s %v = %v\n", featTemp, featTemp.Format(feat), val)
			// }
			retval += t.Mat[i].Value(intTrans, feat)
		}
	}
	return retval
}

func (t *AvgMatrixSparse) SetTransitionScores(features []Feature, scores *[]int64) {
	for i, feat := range features {
		if feat != nil {
			if t.Log {
				featTemp := t.Formatters[i]
				if t.Formatters != nil {
					log.Printf("\t\t%s %v %v\n", featTemp, featTemp.Format(feat), 0)
				}
			}
			t.Mat[i].SetScores(feat, scores)
		}
	}
}

func NewAvgMatrixSparse(features int, formatters []util.Format) *AvgMatrixSparse {
	var (
		Mat []*AvgSparse = make([]*AvgSparse, features)
	)
	for i, _ := range Mat {
		Mat[i] = NewAvgSparse()
	}
	return &AvgMatrixSparse{Mat, features, 0, formatters, AllOut}
}

type AveragedModelStrategy struct {
	P, N       int
	accumModel *AvgMatrixSparse
}

func (u *AveragedModelStrategy) Init(m perceptron.Model, iterations int) {
	// explicitly reset u.N = 0.0 in case of reuse of vector
	// even though 0.0 is zero value
	u.N = 0
	u.P = iterations
	avgModel, ok := m.(*AvgMatrixSparse)
	if !ok {
		panic("AveragedModelStrategy requires AvgMatrixSparse model")
	}
	u.accumModel = avgModel
}

func (u *AveragedModelStrategy) Update(m perceptron.Model) {
	u.accumModel.IncrementGeneration()
	u.N += 1
}

func (u *AveragedModelStrategy) Finalize(m perceptron.Model) perceptron.Model {
	u.accumModel.Generation = u.P * u.N
	u.accumModel.Integrate()
	return u.accumModel
}
