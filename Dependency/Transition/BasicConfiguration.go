package Transition

import (
	"math"
	"regexp"
	"strconv"
)

type SimpleConfiguration struct {
	Stack     Stack
	Queue     Queue
	Arcs      ArcSet
	Nodes     []Token
	Previous  *Configuration
	LastTrans string
}

func (c *SimpleConfiguration) Init(sent Sentence) {
	// Nodes is always the same slice to the same token array
	c.Nodes = sent.([]Token)

	c.Stack = NewStackArray()
	c.Queue = NewQueueSlice(len(sent))
	c.Arcs = NewArcSetSimple()

	for i := int(0); i < len(sent); i++ {
		c.Queue.Enqueue(i)
	}
}

func (c *SimpleConfiguration) Copy() *Configuration {
	newConf := new(SimpleConfiguration)

	newConf.Stack = c.Stack.Copy()
	newConf.Queue = c.Queue.Copy()
	newConf.Arcs = c.Arcs.Copy()

	newConf.Nodes = c.Nodes

	// store a pointer to the previous configuration
	newConf.Previous = c

	return newConf
}

func (c *SimpleConfiguration) SetLastTransition(t string) {
	c.LastTrans = t
}

func (c *SimpleConfiguration) Terminal() bool {
	return c.Queue.Size() == 0
}
