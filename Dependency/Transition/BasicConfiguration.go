package Transition

import (
	"math"
	"regexp"
	"strconv"
)

type BasicConfiguration struct {
	Stack     []int16
	Queue     []int16
	Arcs      []*DepArc
	DepNodes  []*DepNode
	Nodes     []Token
	Previous  *Configuration
	LastTrans string
}

func (c *BasicConfiguration) SetLastTransition(t string) {
	c.LastTrans = t
}

func (c *BasicConfiguration) setNewSent(sent Sentence) {
	for i := int16(0); i < len(sent); i++ {
		c.Queue[i] = i
		leftNodes := make([]int16, 0, 1)
		rightNodes := make([]int16, 0, 1)
		c.DepNodes[i] = &DepNode{-1, leftNodes, rightNodes, i}
	}
}

func (c *BasicConfiguration) Init(sent Sentence) {
	c.Stack = make([]int16, 0, len(sent))
	c.Queue = make([]int16, len(sent))
	c.Arcs = make([]*DepArc, 0, len(sent))
	c.DepNodes = make([]*DepNode, 0, len(sent))
	c.Nodes = sent.([]Token)
	c.setNewSent(sent)
}

func (c *BasicConfiguration) Terminal() bool {
	return len(c.Queue) == 0
}

func (c *BasicConfiguration) Copy() *Configuration {
	newConf := new(BasicConfiguration)

	// ALLOCATION
	newConf.Stack = make([]int16, len(c.Stack))
	newConf.Arcs = make([]*DepArc, len(c.Arcs))
	newConf.DepNodes = make([]*DepNode, len(c.Nodes))
	newConf.Queue = make([]int16, len(c.Queue))

	// DATA COPY
	copy(c.Stack, newConf.Stack)
	copy(c.Queue, newConf.Queue)
	copy(c.DepNodes, newConf.DepNodes)
	copy(c.Arcs, newConf.Arcs)

	// the nodes remain the same, they are a slice
	newConf.Nodes = c.Nodes

	// store a pointer to the previous configuration
	newConf.Previous = c

	return newConf
}
