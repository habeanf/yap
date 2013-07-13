package Transition

import (
	"math"
	"regexp"
	"strconv"
)

type BasicConfiguration struct {
	Stack    []uint16
	Queue    []uint16
	Arcs     []*DepArc
	Nodes    []*DepNode
	Elements []HasProperties
	Previous *Configuration
}

func (c *Configuration) Initialize(initialElements []HasProperties) {
	c.Stack = make([]uint16, 0, len(initialElements))
	c.Queue = make([]uint16, len(initialElements))
	c.Arcs = make([]*DepArc, 0, len(initialElements))
	c.Nodes = make([]*DepNode, 0, len(initialElements))
	c.Elements = initialElements
	for i := uint16(0); i < len(initialElements); i++ {
		c.Queue[i] = i
		leftNodes := make([]uint16, 0, 1)
		rightNodes := make([]uint16, 0, 1)
		c.Nodes[i] = &DepNode{-1, leftNodes, rightNodes, i}
	}
}

func (c *Configuration) Terminal() bool {
	return len(c.Queue) == 0
}

// INCOMPLETE!
func (c *Configuration) Transform(t string) {
	switch t {
	default:
	case "SH":
		c.Stack = append(c.Stack, c.Queue)
		c.Queue = c.Queue[1:]
	case "LA":
	case "RA":
	}
}

func (c *Configuration) Copy() *Configuration {
	newConf := new(Configuration)

	// ALLOCATION
	// the capacity of stack, arcs and nodes is limited to the number
	// tokens, we know the maximum capacity
	newConf.Stack = make([]uint16, len(c.Stack), cap(c.Stack))
	newConf.Arcs = make([]*DepArc, len(c.Arcs), cap(c.Arcs))
	newConf.Nodes = make([]*DepNode, len(c.Nodes), cap(c.Nodes))

	// the queue only gets smaller, no need for all the capacity
	newConf.Queue = make([]uint16, len(c.Queue))

	// DATA COPY
	copy(c.Stack, newConf.Stack)
	copy(c.Queue, newConf.Queue)
	copy(c.Nodes, newConf.Nodes)
	copy(c.Arcs, newConf.Arcs)

	// the elements remain the same, they are a slice
	newConf.Elements = c.Elements

	// store a pointer to the previous configuration
	newConf.Previous = c

	return newConf
}
