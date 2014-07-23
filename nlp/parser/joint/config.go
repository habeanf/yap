package joint

import (
	"chukuparser/alg/transition"
	dep "chukuparser/nlp/parser/dependency/transition"
	"chukuparser/nlp/parser/disambig"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"
)

type JointConfig struct {
	dep.SimpleConfiguration
	disambig.MDConfig

	InternalPrevious *JointConfig
	Last             transition.Transition
}

var _ transition.Configuration = &JointConfig{}
var _ dep.DependencyConfiguration = &JointConfig{}
var _ nlp.DependencyGraph = &JointConfig{}

func (c *JointConfig) Init(abstractLattice interface{}) {
	c.MDConfig.Init(abstractLattice)

	c.Last = 0
	c.InternalPrevious = nil
}

func (c *JointConfig) Terminal() bool {
	return c.MDConfig.Terminal() && c.SimpleConfiguration.Terminal()
}

func (c *JointConfig) Copy() transition.Configuration {
	newConf := new(JointConfig)
	c.CopyTo(newConf)
	return newConf
}

func (c *JointConfig) CopyTo(target transition.Configuration) {
	newConf, ok := target.(*JointConfig)
	if !ok {
		panic("Can't copy into non *JointConfig")
	}

	newConf.Last = c.Last
	newConf.InternalPrevious = c

	c.MDConfig.CopyTo(&newConf.MDConfig)
	c.SimpleConfiguration.CopyTo(&newConf.SimpleConfiguration)
}

func (c *JointConfig) GetSequence() transition.ConfigurationSequence {
	if c.Mappings == nil || c.Arcs() == nil {
		return make(transition.ConfigurationSequence, 0)
	}
	retval := make(transition.ConfigurationSequence, 0, len(c.Mappings)+c.Arcs().Size())
	currentConf := c
	for currentConf != nil {
		retval = append(retval, currentConf)
		currentConf = currentConf.InternalPrevious
	}
	return retval

}

func (c *JointConfig) SetLastTransition(t transition.Transition) {
	c.Last = t
}

func (c *JointConfig) GetLastTransition() transition.Transition {
	return c.Last
}

func (c *JointConfig) String() string {
	return ""
}

func (c *JointConfig) Equal(otherEq util.Equaler) bool {
	if (otherEq == nil && c != nil) || (c == nil && otherEq != nil) {
		return false
	}
	switch other := otherEq.(type) {
	case *JointConfig:
		return (&c.MDConfig).Equal(&other.MDConfig) &&
			(&c.SimpleConfiguration).Equal(&other.SimpleConfiguration)
	}
	panic("Can't equal to non-Joint config")
}

func (c *JointConfig) Address(location []byte, offset int) (nodeID int, exists bool, isGenerator bool) {
	return 0, false, false
}

func (c *JointConfig) GenerateAddresses(nodeID int, location []byte) (nodeIDs []int) {
	return nil
}

func (c *JointConfig) Attribute(source byte, nodeID int, attribute []byte) (interface{}, bool) {
	return nil, false
}

func (c *JointConfig) Previous() transition.Configuration {
	return c.InternalPrevious
}

func (c *JointConfig) Clear() {
	c.SimpleConfiguration.Clear()
	c.MDConfig.Clear()
}
