package Dependency

import "strconv"

type Parser interface {
	Init()
}

type HasProperties interface {
	GetProperty(property string) (string, bool)
}

type DepNode struct {
	HeadIndex    int
	LeftMods     []uint
	RightMods    []uint
	ElementIndex int
}

func (n *DepNode) SetHead(i int) *DepNode {
	newNode := n.Copy()
	newNode.Head = i
	return newNode
}

func (n *DepNode) AttachLeft(i uint) *DepNode {
	newNode := n.Copy()
	newNode.LeftMods = append(newNode.LeftMods, i)
	return newNode
}

func (n *DepNode) AttachRight(i uint) *DepNode {
	newNode := n.Copy()
	newNode.RightMods = append(newNode.RightMods, i)
}

type DepRel string

type DepArc struct {
	Modifier uint
	Head     uint
	Relation DepRel
}

func (arc *DepArc) GetProperty(property string) (string, bool) {
	if property == "l" {
		return arc.Relation, true
	} else {
		return "", false
	}
}

type Configuration struct {
	Stack    []int
	Queue    []int
	Arcs     []*DepArc
	Nodes    []*DepNode
	Elements []HasProperties
}

func (c *Configuration) Transform(t string) {
	switch t {
	default:
	case "SHIFT":
		c.Stack = append(c.Stack, c.Queue)
		c.Queue = c.Queue[1:]
	case "LEFT":
	case "RIGHT":
	}
}

func (c *Configuration) Terminal() bool {
	return len(c.Queue) == 0
}

func (c *Configuration) Initialize(initialElements []HasProperties) {
	c.Stack = make([]int, 0, len(initialElements)/2)
	c.Queue = make([]int, len(initialElements))
	c.Arcs = make([]*DepArc, 0, len(initialElements))
	c.Nodes = make([]*DepNode, 0, len(initialElements))
	c.Elements = initialElements
	for i := 0; i < len(initialElements); i++ {
		c.Queue[i] = i
		leftNodes := make([]uint, 0, len(1))
		rightNodes := make([]uint, 0, len(1))
		c.Nodes[i] = &DepNode{-1, leftNodes, rightNodes, i}
	}
}

func (c *Configuration) Copy() *Configuration {
	newConf := new(Configuration)

	// ALLOCATION
	// the capacity of stack, arcs and nodes is limited to the number
	// tokens, we know the maximum capacity
	newConf.Stack = make([]uint, len(c.Stack), cap(c.Stack))
	newConf.Arcs = make([]*DepArc, len(c.Arcs), cap(c.Arcs))
	newConf.Nodes = make([]*DepNode, len(c.Nodes), cap(c.Nodes))

	// the queue only gets smaller, no need for all the capacity
	newConf.Queue = make([]uint, len(c.Queue))

	// DATA COPY
	copy(c.Stack, newConf.Stack)
	copy(c.Queue, newConf.Queue)
	copy(c.Nodes, newConf.Nodes)
	copy(c.Arcs, newConf.Arcs)

	// the elements remain the same, they are a slice
	newConf.Elements = c.Elements

	return newConf
}

func (c *Configuration) GetProperty(property string) (string, bool) {
	if property == "d" {
		return "1", true
	} else {
		return "", false
	}
}

func (c *Configuration) GetSource(source string) *interface{} {
	switch source {
	case "N":
		return &(c.Queue)
	case "S":
		return &(c.Stack)
	}
	return nil
}

func (c *Configuration) GetLocation(currentTarget interface{}, location []byte) (*HasProperties, bool) {
	switch t := currentTarget.(type) {
	default:
		return nil, false
	case *[]DepNode:
		return c.GetLocationNodeStack(t, location)
	case *DepNode:
		return c.GetLocationDepNode(t, location)
	case *DepArc:
		// currentTarget is a DepArc
		// location remainder is discarded
		// (currently no navigation on the arc)
		return currentTarget.(*HasProperties), true
	}
}

func (c *Configuration) GetLocationNodeStack(stack *[]DepNode, location string) (*HasProperties, bool) {
	// currentTarget is a slice
	// location "head" must be an offset
	offset, err := strconv.ParseInt(currentLocation, 10, 0)
	if !err {
		panic("Error parsing location string " + location + " ; " + err.Error())
	}
	// if a referenced location cannot exist
	// return an empty result
	if len(t) <= offset {
		return nil, false
	}
	return c.GetLocation(stack, location[len(currentLocation):])
}

func (c *Configuration) GetLocationDepNode(node *DepNode, location string) (*HasProperties, bool) {
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
				return head.(*HasProperties), true
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

func (c *Configuration) GetDepNodeHead(node *DepNode) (*DepNode, bool) {
	if node.HeadIndex == -1 {
		return nil, false
	}
	if len(c.Nodes) <= node.HeadIndex {
		panic("Node referenced head out of bounds")
	}
	return c.Nodes[node.HeadIndex], true
}
