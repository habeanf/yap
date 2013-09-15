package Transition

import . "chukuparser/Algorithm/FeatureVector"

type Model interface {
	TransitionScore(transitionID int, features []Feature) float64
}
