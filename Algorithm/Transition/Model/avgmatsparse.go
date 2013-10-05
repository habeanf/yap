package Model

import (
	. "chukuparser/Algorithm/FeatureVector"
	"chukuparser/Algorithm/Perceptron"
	"chukuparser/Algorithm/Transition"
	"chukuparser/Util"
	"log"
)

var allOut bool = false

type AvgMatrixSparse struct {
	Mat                  []AvgSparse
	Features, Generation int
	Formatters           []Util.Format
	Log                  bool
}

var _ Perceptron.Model = &AvgMatrixSparse{}
var _ Interface = &AvgMatrixSparse{}

func (t *AvgMatrixSparse) Score(features interface{}) float64 {
	var (
		retval   float64
		intTrans int
	)
	f := features.(*FeaturesList)
	if f.Previous == nil {
		return 0
	}
	lastTransition := f.Transition
	featuresList := f.Previous
	for featuresList != nil {
		intTrans = int(lastTransition)
		for i, feature := range featuresList.Features {
			if feature != nil {
				retval += t.Mat[i].Value(intTrans, feature)
			}
		}
		lastTransition = featuresList.Transition
		featuresList = featuresList.Previous
	}
	return retval
}

func (t *AvgMatrixSparse) Add(features interface{}) Perceptron.Model {
	if t.Log {
		log.Println("Score", 1.0, "to")
	}
	t.apply(features, 1.0)
	return t
}

func (t *AvgMatrixSparse) Subtract(features interface{}) Perceptron.Model {
	if t.Log {
		log.Println("Score", -1.0, "to")
	}
	t.apply(features, -1.0)
	return t
}

func (t *AvgMatrixSparse) apply(features interface{}, amount float64) Perceptron.Model {
	var (
		intTrans int
	)
	f := features.(*FeaturesList)
	if f.Previous == nil {
		return t
	}
	lastTransition := f.Transition
	featuresList := f.Previous
	t.apply(f.Previous, amount)
	// for featuresList != nil {
	intTrans = int(lastTransition)
	if t.Log {
		log.Println("\tstate", intTrans)
	}
	for i, feature := range featuresList.Features {
		if feature != nil {
			if t.Log {
				featTemp := t.Formatters[i]
				if t.Formatters != nil && i < 60 {
					log.Printf("\t\t%s %v %v\n", featTemp, featTemp.Format(feature), amount)
				}
			}
			t.Mat[i].Add(t.Generation, intTrans, feature, amount)
		}
	}
	// 	lastTransition = featuresList.Transition
	// 	featuresList = featuresList.Previous
	// }
	return t
}

func (t *AvgMatrixSparse) ScalarDivide(val float64) {
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

func (t *AvgMatrixSparse) Copy() Perceptron.Model {
	panic("Cannot copy an avg matrix sparse representation")
	return nil
}

func (t *AvgMatrixSparse) New() Perceptron.Model {
	return NewAvgMatrixSparse(t.Features, nil)
}

func (t *AvgMatrixSparse) AddModel(m Perceptron.Model) {
	panic("Cannot add two avg matrix sparse types")
}

func (t *AvgMatrixSparse) TransitionScore(transition Transition.Transition, features []Feature) float64 {
	var (
		retval   float64
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

func (t *AvgMatrixSparse) SetTransitionScores(features []Feature, scores *[]float64) {
	for i, feat := range features {
		if feat != nil {
			t.Mat[i].SetScores(feat, scores)
		}
	}
}

func NewAvgMatrixSparse(features int, formatters []Util.Format) *AvgMatrixSparse {
	var (
		Mat []AvgSparse = make([]AvgSparse, features)
	)
	for i, _ := range Mat {
		Mat[i] = make(AvgSparse)
	}
	return &AvgMatrixSparse{Mat, features, 0, formatters, allOut}
}

type AveragedModelStrategy struct {
	P, N       int
	accumModel *AvgMatrixSparse
}

func (u *AveragedModelStrategy) Init(m Perceptron.Model, iterations int) {
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

func (u *AveragedModelStrategy) Update(m Perceptron.Model) {
	u.accumModel.IncrementGeneration()
	u.N += 1
}

func (u *AveragedModelStrategy) Finalize(m Perceptron.Model) Perceptron.Model {
	u.accumModel.Generation = u.P * u.N
	u.accumModel.Integrate()
	return u.accumModel
}
