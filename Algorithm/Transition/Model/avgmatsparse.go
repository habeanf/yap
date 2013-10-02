package Model

import (
	. "chukuparser/Algorithm/FeatureVector"
	"chukuparser/Algorithm/Perceptron"
	"chukuparser/Algorithm/Transition"
	"log"
)

type AvgMatrixSparse struct {
	Mat                   [][]AvgSparse
	Transitions, Features int
	Generation            int
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
		if intTrans < t.Transitions {
			for i, feature := range featuresList.Features {
				if feature != nil {
					retval += t.Mat[intTrans][i].Value(feature)
				}
			}
		}
		lastTransition = featuresList.Transition
		featuresList = featuresList.Previous
	}
	return retval
}

func (t *AvgMatrixSparse) Add(features interface{}) Perceptron.Model {
	var (
		intTrans int
	)
	f := features.(*FeaturesList)
	if f.Previous == nil {
		return t
	}
	log.Println("Score 1 to")
	lastTransition := f.Transition
	featuresList := f.Previous
	for featuresList != nil {
		intTrans = int(lastTransition)
		log.Println("\tstate", intTrans)
		if intTrans >= t.Transitions {
			t.ExtendTransitions(intTrans)
		}
		for i, feature := range featuresList.Features {
			if feature != nil {
				t.Mat[intTrans][i].Increment(t.Generation, feature)
			}
		}
		lastTransition = featuresList.Transition
		featuresList = featuresList.Previous
	}
	return t
}

func (t *AvgMatrixSparse) Subtract(features interface{}) Perceptron.Model {
	var (
		intTrans int
	)
	f := features.(*FeaturesList)
	if f.Previous == nil {
		return t
	}
	log.Println("Score -1 to")
	lastTransition := f.Transition
	featuresList := f.Previous
	for featuresList != nil {
		intTrans = int(lastTransition)
		log.Println("\tstate", intTrans)
		if intTrans >= t.Transitions {
			t.ExtendTransitions(intTrans)
		}
		for i, feature := range featuresList.Features {
			if feature != nil {
				t.Mat[intTrans][i].Decrement(t.Generation, feature)
			}
		}
		lastTransition = featuresList.Transition
		featuresList = featuresList.Previous
	}
	return t
}

func (t *AvgMatrixSparse) ScalarDivide(val float64) {
	for i, _ := range t.Mat {
		for _, avgsparse := range t.Mat[i] {
			avgsparse.UpdateScalarDivide(val)
		}
	}
}

func (t *AvgMatrixSparse) Integrate() {
	for i, _ := range t.Mat {
		for j, _ := range t.Mat[i] {
			t.Mat[i][j].Integrate(t.Generation)
		}
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
	return NewAvgMatrixSparse(t.Transitions, t.Features)
}

func (t *AvgMatrixSparse) AddModel(m Perceptron.Model) {
	panic("Cannot add two avg matrix sparse types")
}

func (t *AvgMatrixSparse) TransitionScore(transition Transition.Transition, features []Feature) float64 {
	var (
		retval   float64
		intTrans int = int(transition)
	)
	if intTrans >= t.Transitions {
		return 0.0
	}
	if intTrans < 0 {
		panic("Got negative transition index")
	}
	featuresArray := t.Mat[intTrans]

	for i, feat := range features {
		if feat != nil {
			retval += featuresArray[i].Value(feat)
		}
	}
	return retval
}

func (t *AvgMatrixSparse) ExtendTransitions(extendTo int) {
	newTransitions := extendTo - t.Transitions + 1
	ExtraMat1D := make([]AvgSparse, newTransitions*t.Features)
	for i := range ExtraMat1D {
		ExtraMat1D[i] = make(AvgSparse)
	}
	for i := 0; i < newTransitions; i++ {
		t.Mat, ExtraMat1D = append(t.Mat, ExtraMat1D[:t.Features]), ExtraMat1D[t.Features:]
	}
	t.Transitions = extendTo + 1
}

func NewAvgMatrixSparse(transitions, features int) *AvgMatrixSparse {
	var (
		Mat1D []AvgSparse   = make([]AvgSparse, transitions*features)
		Mat2D [][]AvgSparse = make([][]AvgSparse, 0, transitions)
	)
	for i, _ := range Mat1D {
		Mat1D[i] = make(AvgSparse)
	}
	for i := 0; i < transitions; i++ {
		Mat2D, Mat1D = append(Mat2D, Mat1D[:features]), Mat1D[features:]
	}
	return &AvgMatrixSparse{Mat2D, transitions, features, 0}
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
