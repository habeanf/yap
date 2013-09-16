package Model

import (
	. "chukuparser/Algorithm/FeatureVector"
	"chukuparser/Algorithm/Perceptron"
	"chukuparser/Algorithm/Transition"
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
	featuresList := features.(*FeaturesList)
	for featuresList != nil {
		intTrans = int(featuresList.Transition)
		if intTrans < t.Transitions {
			for i, feature := range featuresList.Features {
				retval += t.Mat[int(featuresList.Transition)][i].Value(feature)
			}
		}
		featuresList = featuresList.Previous
	}
	return retval
}

func (t *AvgMatrixSparse) Add(features interface{}) Perceptron.Model {
	var (
		intTrans int
	)
	featuresList := features.(*FeaturesList)
	for featuresList != nil {
		intTrans = int(featuresList.Transition)
		if intTrans >= t.Transitions {
			t.ExtendTransitions(intTrans)
		}
		for i, feature := range featuresList.Features {
			t.Mat[intTrans][i].Increment(t.Generation, feature)
		}
		featuresList = featuresList.Previous
	}
	return t
}

func (t *AvgMatrixSparse) Subtract(features interface{}) Perceptron.Model {
	var (
		intTrans int
		exists   bool
	)
	featuresList := features.(*FeaturesList)
	for featuresList != nil {
		intTrans = int(featuresList.Transition)
		if intTrans >= t.Transitions {
			t.ExtendTransitions(intTrans)
		}
		for i, feature := range featuresList.Features {
			t.Mat[intTrans][i].Decrement(t.Generation, feature)
		}
		featuresList = featuresList.Previous
	}
	return t
}

func (t *AvgMatrixSparse) ScalarDivide(val float64) {
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
		retval += featuresArray[i].Value(feat)
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
	return &MatrixSparse{Mat2D, transitions, features}
}
