package Model

// import (
// 	. "chukuparser/algorithm/featurevector"
// 	"chukuparser/algorithm/perceptron"
// 	"chukuparser/algorithm/transition"
// )

// type Trivial struct {
// 	Vec Sparse
// }

// var _ Perceptron.Model = &Trivial{}
// var _ Interface = &Trivial{}

// func (t *Trivial) Score(features interface{}) float64 {
// 	var (
// 		retval float64
// 		feat   interface{}
// 	)
// 	featuresList := features.(*Transition.FeaturesList)
// 	for featuresList != nil {
// 		for i, feature := range featuresList.Features {
// 			feat = MakeFeature(int(featuresList.Transition), i, feature)
// 			retval += t.Vec[feat]
// 		}
// 		featuresList = featuresList.Previous
// 	}
// 	return retval
// }

// func (t *Trivial) Add(features interface{}) Perceptron.Model {
// 	var (
// 		curval float64
// 		feat   interface{}
// 	)
// 	featuresList := features.(*Transition.FeaturesList)
// 	for featuresList != nil {
// 		for i, feature := range featuresList.Features {
// 			feat = MakeFeature(int(featuresList.Transition), i, feature)
// 			curval, _ = t.Vec[feat]
// 			t.Vec[feat] = curval + 1
// 		}
// 		featuresList = featuresList.Previous
// 	}
// 	return t
// }

// func (t *Trivial) Subtract(features interface{}) Perceptron.Model {
// 	var (
// 		curval float64
// 		feat   interface{}
// 	)
// 	featuresList := features.(*Transition.FeaturesList)
// 	for featuresList != nil {
// 		for i, feature := range featuresList.Features {
// 			feat = MakeFeature(int(featuresList.Transition), i, feature)
// 			curval, _ = t.Vec[feat]
// 			t.Vec[feat] = curval - 1
// 		}
// 		featuresList = featuresList.Previous
// 	}
// 	return t
// }

// func (t *Trivial) ScalarDivide(val float64) {
// 	t.Vec.UpdateScalarDivide(val)
// }

// func (t *Trivial) Copy() Perceptron.Model {
// 	return &Trivial{t.Vec.Copy()}
// }

// func (t *Trivial) New() Perceptron.Model {
// 	return NewTrivial()
// }

// func (t *Trivial) AddModel(m Perceptron.Model) {
// 	other, ok := m.(*Trivial)
// 	if !ok {
// 		panic("Can't add perceptron model not of the same type")
// 	}
// 	t.Vec.Add(other.Vec)
// }

// func (t *Trivial) TransitionScore(transition Transition.Transition, features []Feature) float64 {
// 	var (
// 		retval float64
// 	)

// 	for i, feat := range features {
// 		retval += t.Vec[MakeFeature(int(transition), i, feat)]
// 	}
// 	return retval
// }

// func NewTrivial() *Trivial {
// 	return &Trivial{NewSparse()}
// }
