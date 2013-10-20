package transition

import (
// "log"
// nlp "chukuparser/nlp/types"
// "chukuparser/util"
// "math"
// "regexp"
// "sort"
// "strconv"
// "strings"
)

const (
	SET_SEPARATOR = "-"
)

func (c *SimpleConfiguration) Address(location []byte, sourceOffset int) (int, bool) {
	source := c.GetSource(location[0])
	if source == nil {
		return 0, false
	}
	atAddress, exists := source.Index(int(sourceOffset))
	if !exists {
		// zpar bug parity
		if location[0] == 'N' && c.Queue().Size() == 0 && sourceOffset > 0 {
			return sourceOffset - 1, true
		}
		// end zpar bug parity
		return 0, false
	}
	location = location[2:]
	if len(location) == 0 {
		return atAddress, true
	}
	switch location[0] {
	case 'l', 'r':
		leftMods, rightMods := c.GetModifiers(atAddress)
		if location[0] == 'l' {
			if len(leftMods) == 0 {
				return 0, false
			}
			if len(location) > 1 && location[1] == '2' {
				if len(leftMods) > 1 {
					return leftMods[1], true
				}
			} else {
				return leftMods[0], true
			}
		} else {
			if len(rightMods) == 0 {
				return 0, false
			}
			if len(location) > 1 && location[1] == '2' {
				if len(rightMods) > 1 {
					return rightMods[len(rightMods)-2], true
				}
			} else {
				return rightMods[len(rightMods)-1], true
			}
		}
	case 'h':
		head, headExists := c.GetHead(atAddress)
		if headExists {
			if len(location) > 1 && location[1] == '2' {
				headOfHead, headOfHeadExists := c.GetHead(head.ID())
				if headOfHeadExists {
					return headOfHead.ID(), true
				}
			} else {
				return head.ID(), true
			}
		}
	}
	return 0, false
}

func (c *SimpleConfiguration) GetModifierLabel(modifierID int) (int, bool) {
	arcs := c.Arcs().Get(&BasicDepArc{-1, -1, modifierID, ""})
	if len(arcs) > 0 {
		index, _ := c.ERel.IndexOf(arcs[0].GetRelation())
		return index, true
	}
	return 0, false
}

func (c *SimpleConfiguration) Attribute(source byte, nodeID int, attribute []byte) (interface{}, bool) {
	if nodeID < 0 || nodeID >= len(c.Nodes) {
		return 0, false
	}
	switch attribute[0] {
	case 'd':
		return c.GetConfDistance()
	case 'w':
		node := c.GetRawNode(nodeID)
		return node.Token, true
	case 'p':
		node := c.GetRawNode(nodeID)
		// TODO: CPOS
		return node.POS, true
	case 'l':
		//		relation, relExists :=
		return c.GetModifierLabel(nodeID)
	case 'v':
		if len(attribute) != 2 {
			return 0, false
		}
		leftMods, rightMods := c.GetNumModifiers(nodeID)
		switch attribute[1] {
		case 'l':
			return leftMods, true
		case 'r':
			return rightMods, true
		}
	case 's':
		if len(attribute) != 2 {
			return 0, false
		}
		leftLabelSet, rightLabelSet := c.GetModifierLabelSets(nodeID)
		switch attribute[1] {
		case 'l':
			return leftLabelSet, true
		case 'r':
			return rightLabelSet, true
		}
	}
	return 0, false
}

func (c *SimpleConfiguration) GetConfDistance() (int, bool) {
	stackTop, stackExists := c.Stack().Peek()
	queueTop, queueExists := c.Queue().Peek()
	if stackExists && queueExists {
		dist := queueTop - stackTop
		// "normalize" to
		// 0 1 2 3 4 5 ... 10 ...
		// 0 1 2 3 4 ---5--  --- 6 ---
		if dist < 0 {
			dist = -dist
		}
		switch {
		case dist > 10:
			return 6, true
		case dist > 5:
			return 5, true
		default:
			return dist, true
		}
	}
	return 0, false
}

func (c *SimpleConfiguration) GetSource(location byte) Index {
	switch location {
	case 'N':
		return c.Queue()
	case 'S':
		return c.Stack()
	}
	return nil
}

func (c *SimpleConfiguration) GetHead(nodeID int) (*ArcCachedDepNode, bool) {
	head := c.Nodes[nodeID].Head
	if head == -1 {
		return nil, false
	}
	return c.Nodes[head], true
}

func (c *SimpleConfiguration) GetModifiers(nodeID int) ([]int, []int) {
	node := c.Nodes[nodeID]
	return node.LeftMods(), node.RightMods()
}

func (c *SimpleConfiguration) GetNumModifiers(nodeID int) (int, int) {
	node := c.Nodes[nodeID]
	return len(node.LeftMods()), len(node.RightMods())
}

func (c *SimpleConfiguration) GetModifierLabelSets(nodeID int) (interface{}, interface{}) {
	node := c.Nodes[nodeID]
	return node.LeftLabelSet(), node.RightLabelSet()
}
