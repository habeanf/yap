package Perceptron

import (
	. "chukuparser/Algorithm/FeatureVector"
	"chukuparser/Util"
)

type Model interface {
	Util.Persist
	Score(features []Feature) float64
	Add(features []Feature) Model
	Subtract(features []Feature) Model
	ScalarDivide(float64)
}

type Instance interface {
	Util.Equaler
}

type DecodedInstance interface {
	Instance
	Instance() Instance
	Decoded() interface{}
}

type Decoded struct {
	InstanceVal Instance
	DecodedVal  Util.Equaler
}

var _ DecodedInstance = &Decoded{}

func (d *Decoded) Decoded() interface{} {
	return d.DecodedVal
}

func (d *Decoded) Instance() Instance {
	return d.InstanceVal
}

func (d *Decoded) Equal(otherEq Util.Equaler) bool {
	if otherEq == nil {
		return false
	}
	other := otherEq.(*Decoded)
	instanceEq := d.InstanceVal.Equal(other.InstanceVal)
	decodedEq := d.DecodedVal.Equal(other.DecodedVal)
	return instanceEq && decodedEq
}

type FeatureExtractor interface {
	Features(Instance) []Feature
	EstimatedNumberOfFeatures() int
}

type InstanceDecoder interface {
	Decode(i Instance, m Model) (DecodedInstance, []Feature)
	DecodeGold(i DecodedInstance, m Model) (DecodedInstance, []Feature)
}

type EarlyUpdateInstanceDecoder interface {
	DecodeEarlyUpdate(i DecodedInstance, m Model) (decoded DecodedInstance, decodedFeatures, goldFeatures []Feature)
}

type SupervisedTrainer interface {
	Train(instances []DecodedInstance)
}

// unused, here for completeness
type UnsupervisedTrainer interface {
	Train(instances []Instance)
}
