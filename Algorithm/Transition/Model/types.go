package Model

import (
	. "chukuparser/Algorithm/FeatureVector"
	"chukuparser/Algorithm/Model/Perceptron"
)

type Interface interface {
	TransitionScore(transitionID int, features []Feature) float64
}
