package Algorithm

type Model interface {
	Score(i *DecodedInstance) float64
}

type Instance interface{}

type DecodedInstance interface {
	Instance
	Decode() interface{}
	SetInstance(i *Instance)
	Equals(other *DecodedInstance) bool
}

type Feature string

type FeatureExtractor interface {
	Features(Instance) *[]Feature
	NumberOfFeatures() int
}

type Classifier interface {
	Classify(instance *Instance) *DecodedInstance
}

type Trainer interface {
	Train(instances chan *Instance)
}

type SupervisedTrainer interface {
	Train(instances chan *DecodedInstance)
}

type UnsupervisedTrainer interface {
	Train(instances chan *Instance)
}

type Decoder interface {
	Decode(i *Instance, m *Model) *DecodedInstance
}
