package Perceptron

import "chukuparser/Util"

type Model interface {
	Score(features []Feature) float64
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
	other := otherEq.(*Decoded)
	instanceEq := d.InstanceVal.Equal(other.InstanceVal)
	decodedEq := d.DecodedVal.Equal(other.DecodedVal)
	return instanceEq && decodedEq
}

type Feature string

type FeatureExtractor interface {
	Features(Instance) []Feature
	EstimatedNumberOfFeatures() int
}

type InstanceDecoder interface {
	Decode(i Instance, m Model) (DecodedInstance, *SparseWeightVector)
	DecodeGold(i DecodedInstance, m Model) (DecodedInstance, *SparseWeightVector)
}

type EarlyUpdateInstanceDecoder interface {
	DecodeEarlyUpdate(i DecodedInstance, m Model) (DecodedInstance, *SparseWeightVector, *SparseWeightVector)
}

type SupervisedTrainer interface {
	Train(instances []DecodedInstance)
}

// unused, here for completeness
type UnsupervisedTrainer interface {
	Train(instances []Instance)
}
