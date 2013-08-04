package Transition

import (
	. "chukuparser/Algorithm/Model/Perceptron"
	. "chukuparser/NLP"

	// "math"
	// "regexp"
	"strconv"
)

// Code copied from float64 version in math/abs.go
func absInt(x int) int {
	switch {
	case x < 0:
		return -x
	case x == 0:
		return 0 // return correctly abs(-0)
	}
	return x
}

func (c SimpleConfiguration) Attribute(attr string) (string, bool) {
	if attr != "d" {
		return "", false
	}
	stackTop, stackExists := c.Stack().Peek()
	queueTop, queueExists := c.Queue().Peek()
	if stackExists && queueExists {
		return string(absInt(stackTop - queueTop)), true
	}
	return "", false
}

func (c SimpleConfiguration) Address(location []byte) (*interface{}, bool) {
	switch source {
	case "N":
		q := (Addressable)(c.Queue())
		return &q
	case "S":
		s := (Addressable)(c.Stack())
		return &s
	}
	return nil
}

func (c SimpleConfiguration) Attribute(currentTarget *interface{}, location []byte) (*Attributes, bool) {
	target := *currentTarget
	switch t := target.(type) {
	case DepNode:
		return c.getNodeAttribute(t, location)
	case LabeledDepArc:
		// currentTarget is a DepArc
		// location remainder is discarded
		// (currently no navigation on the arc)
		return t.getArcAttribute(t, location), true
	case SimpleConfiguration:
		return t.getAttribute(t, location)
	default:
		return nil, false
	}
}

func (c SimpleConfiguration) GetAddressNodeStack(stack *[]int, location []byte) (*Attributes, bool) {
	// currentTarget is a slice
	currentLocation := location[0]
	// location "head" must be an offset
	offset, err := strconv.ParseInt(string(currentLocation), 10, 0)
	if err != nil {
		panic("Error parsing location string " + string(currentLocation) + " ; " + err.Error())
	}
	// if a referenced location cannot exist
	// return an empty result
	if len(*stack) <= int(offset) {
		return nil, false
	}
	stackLocation := (*stack)[offset]
	return c.GetAddress(&stackLocation, location[len(currentLocation):])
}

// INCOMPLETE!
func (c SimpleConfiguration) GetAddressDepNode(node *DepNode, location []byte) (*Attributes, bool) {
	// location "head" can be either:
	// - empty (return the currentTarget)
	// - the leftmost/rightmost (l/r) arc
	// - the k-th leftmost/rightmost (lNNN/rNNN) arc
	// - the head (h)
	// - the k-th head (hNNN)
	if len(location) == 0 {
		return node, true
	}
	locationRemainder := location[1:]
	switch location[0] {
	case "h":
		if len(locationRemainder) == 0 {
			if head, exists := c.GetDepNodeHead(node); exists {
				return head.(*Attributes), true
			} else {
				return nil, false
			}
		}
		return
	case "l":
		return
	case "r":
		return
	}

}

func (c SimpleConfiguration) GetDepNodeHead(node *DepNode) (*DepNode, bool) {
	if node.HeadIndex == -1 {
		return nil, false
	}
	if len(c.Nodes) <= node.HeadIndex {
		panic("Node referenced head out of bounds")
	}
	return c.Nodes[node.HeadIndex], true
}
