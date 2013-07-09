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
