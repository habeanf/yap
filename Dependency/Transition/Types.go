package Transition

type HasProperties interface {
	GetProperty(property string) (string, bool)
}

type DepNode struct {
	HeadIndex    int16
	LeftMods     []uint16
	RightMods    []uint16
	ElementIndex uint16
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

type Stack struct {
	Value    uint16
	Previous *Stack
}

func (s *Stack) Pop() (*Stack, uint16, bool) {
	if s == nil {
		return nil, 0, false
	}
	retval := s.Value
	return s.Previous, retval, true
}

func (s *Stack) Push(value int) *Stack {
	newStackHead := new(Stack)
	newStackHead.Value = value
	newStackHead.Previous = s
	return newStackHead
}

func (s *Stack) GetDepth(depth int) uint16, bool{
	current := s
	for depth>0; depth-- {
		if current.Previous == nil {
			return 0, false
		}
		current = current.Previous
	}
	return current.Value, true
}