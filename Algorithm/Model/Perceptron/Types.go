package Model

type Model interface {
	Score(i *DecodedInstance) float64
}

type Instance interface {
	ID() int
}

type DecodedInstance interface {
	Instance
	Decode() interface{}
	Equals(other *DecodedInstance) bool
	GetInstance() *Instance
}

type Feature string

type FeatureExtractor interface {
	Features(*Instance) *[]Feature
	EstimatedNumberOfFeatures() int
}

type Classifier interface {
	Classify(instance *Instance) *DecodedInstance
}

type Trainer interface {
	Train(instances chan *Instance)
}

type SupervisedTrain interface {
	Train(instances chan *DecodedInstance)
}

type UnsupervisedTrain interface {
	Train(instances chan *Instance)
}

type Decoder interface {
	Decode(i *Instance, m Model) *DecodedInstance
}

type HasAttributes interface {
	GetProperty(property string) (string, bool)
}
