package disambig

import (
	// "chukuparser/algorithm/graph"
	. "chukuparser/algorithm"
	. "chukuparser/algorithm/transition"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"

	// "fmt"
	// "log"
	// "reflect"
	// "strings"
)

type MDConfig struct {
	LatticeQueue Queue
	Lattices     []nlp.Lattice
	Mappings     nlp.Mappings

	InternalPrevious *MDConfig
	Last             Transition
}

var _ Configuration = &MDConfig{}

func (c *MDConfig) Init(abstractLattice interface{}) {
	latticeSent := abstractLattice.(nlp.LatticeSentence)
	sentLength := len(latticeSent)

	c.Lattices = latticeSent

	maxSentLength := 0
	var latP *nlp.Lattice
	for _, lat := range c.Lattices {
		latP = &lat
		maxSentLength += latP.MaxPathLen()
	}

	c.LatticeQueue = NewQueueSlice(sentLength)
	c.Mappings = make([]*nlp.Mapping, 0, len(c.Lattices))

	// push indexes of statement nodes to *LatticeQueue*, in reverse order (first word at the top of the queue)
	for i := 0; i < sentLength; i++ {
		c.LatticeQueue.Enqueue(i)
	}

	// explicit resetting of zero-valued properties
	// in case of reuse
	c.Last = 0
}

func (c *MDConfig) Terminal() bool {
	return c.LatticeQueue.Size() == 0
}

func (c *MDConfig) Copy() Configuration {
	newConf := new(MDConfig)
	newConf.Mappings = make([]*nlp.Mapping, len(c.Mappings), len(c.Lattices))
	copy(newConf.Mappings, c.Mappings)

	if c.LatticeQueue != nil {
		newConf.LatticeQueue = c.LatticeQueue.Copy()
	}

	// lattices slice is read only, no need for copy
	newConf.Lattices = c.Lattices
	newConf.InternalPrevious = c
	return newConf
}

func (c *MDConfig) GetSequence() ConfigurationSequence {
	if c.Mappings == nil {
		return make(ConfigurationSequence, 0)
	}
	retval := make(ConfigurationSequence, 0, len(c.Mappings))
	currentConf := c
	for currentConf != nil {
		retval = append(retval, currentConf)
		currentConf = currentConf.InternalPrevious
	}
	return retval
}

func (c *MDConfig) SetLastTransition(t Transition) {
	c.Last = t
}

func (c *MDConfig) GetLastTransition() Transition {
	return Transition(c.Last)
}

func (c *MDConfig) String() string {
	return "TODO: Implement String()"
}

func (c *MDConfig) Equal(otherEq util.Equaler) bool {
	if (otherEq == nil && c != nil) || (c == nil && otherEq != nil) {
		return false
	}
	switch other := otherEq.(type) {
	case *MDConfig:
		if (other == nil && c != nil) || (c == nil && other != nil) {
			return false
		}
		if other.Last != c.Last {
			return false
		}
		if c.InternalPrevious == nil && other.InternalPrevious == nil {
			return true
		}
		if c.InternalPrevious != nil && other.InternalPrevious != nil {
			return c.InternalPrevious.Equal(other.InternalPrevious)
		} else {
			return false
		}
	default:
		panic("TODO: Figure out what the type of the other is ([]*nlp.Mapping?)")
	}
}

func (c *MDConfig) Previous() *MDConfig {
	return c.InternalPrevious
}

func (c *MDConfig) Clear() {

}
