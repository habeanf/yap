package Perceptron

type Model interface {
	Score(i DecodedInstance) float64
}

type Instance interface {
	ID() int
}

type DecodedInstance interface {
	Instance
	Decode() Equal
	Equal(other *DecodedInstance) bool
	GetInstance() Instance
}

type Decoded struct {
	instance Instance
	decoded  Equal
}

var _ DecodedInstance = &Decoded{}

func (d *Decoded) Decode() interface{} {
	return d.decoded
}

func (d *Decoded) GetInstance() Instance {
	return d.instance
}

func (d *Decoded) Equals(other *DecodedInstance) bool {
	return instance.ID() == other.ID() && d.Decode().Equal(other.Decode())
}

type Feature string

type FeatureExtractor interface {
	Features(Instance) []Feature
	EstimatedNumberOfFeatures() int
}

type InstanceDecoder interface {
	Decode(i Instance, m Model) DecodedInstance
}

type SupervisedTrainer interface {
	Train(instances chan DecodedInstance)
}

// unused, here for completeness
type UnsupervisedTrainer interface {
	Train(instances chan Instance)
}
