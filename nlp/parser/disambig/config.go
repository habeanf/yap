package disambig

import (
	// "chukuparser/algorithm/graph"
	. "chukuparser/algorithm"
	. "chukuparser/algorithm/transition"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"

	"fmt"
	// "log"
	// "reflect"
	"strings"
)

type MDConfig struct {
	LatticeQueue Queue
	Lattices     nlp.LatticeSentence
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
	if c.Mappings == nil {
		return fmt.Sprintf("\t=>([],\t[])")
	}
	mapLen := len(c.Mappings)
	if mapLen > 0 {
		return fmt.Sprintf("MD\t=>([%s],\t[%v])", c.StringLatticeQueue(), c.Mappings[mapLen-1])
	} else {
		return fmt.Sprintf("\t=>([%s],\t[%s])", c.StringLatticeQueue(), "")
	}
}

func (c *MDConfig) StringLatticeQueue() string {
	queueSize := c.LatticeQueue.Size()
	switch {
	case queueSize > 0 && queueSize <= 3:
		var queueStrings []string = make([]string, 0, 3)
		for i := 0; i < c.LatticeQueue.Size(); i++ {
			atI, _ := c.LatticeQueue.Index(i)
			queueStrings = append(queueStrings, string(c.Lattices[atI].Token))
		}
		return strings.Join(queueStrings, ",")
	case queueSize > 3:
		headID, _ := c.LatticeQueue.Index(0)
		tailID, _ := c.LatticeQueue.Index(c.LatticeQueue.Size() - 1)
		head := c.Lattices[headID]
		tail := c.Lattices[tailID]
		return strings.Join([]string{string(head.Token), "...", string(tail.Token)}, ",")
	default:
		return ""
	}

}
func (c *MDConfig) Equal(otherEq util.Equaler) bool {
	if (otherEq == nil && c != nil) || (c == nil && otherEq != nil) {
		// log.Println("\tfalse default")
		return false
	}
	switch other := otherEq.(type) {
	case *MDConfig:
		if (other == nil && c != nil) || (c == nil && other != nil) {
			// log.Println("\tfalse 0")
			return false
		}
		// log.Println("Comparing", c, "to", other)
		// log.Println("Comparing\n", c.GetSequence(), "\n\tto\n", other.GetSequence())
		if other.Last != c.Last {
			// log.Println("\tfalse 1")
			return false
		}
		if c.InternalPrevious == nil && other.InternalPrevious == nil {
			// log.Println("\ttrue")
			return true
		}
		if c.InternalPrevious != nil && other.InternalPrevious != nil {
			// log.Println("\trecurse")
			return c.InternalPrevious.Equal(other.InternalPrevious)
		} else {
			// log.Println("\tfalse 3: ", c.InternalPrevious, "vs", other.InternalPrevious)
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
	c.InternalPrevious = nil
}

func (c *MDConfig) Address(location []byte, sourceOffset int) (int, bool, bool) {
	return 0, false, false
}

func (c *MDConfig) Attribute(source byte, nodeID int, attribute []byte) (interface{}, bool) {
	return nil, false
}

func (c *MDConfig) GenerateAddresses(nodeID int, location []byte) (nodeIDs []int) {
	return nil
}
