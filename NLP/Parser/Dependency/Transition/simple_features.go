package Transition

import (
	"chukuparser/Util"
	// "math"
	// "regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	SET_SEPARATOR = "-"
)

func (c *SimpleConfiguration) Address(location []byte) (int, bool) {
	source := c.getSource(location[0])
	if source == nil {
		return 0, false
	}
	sourceOffset, err := strconv.ParseInt(string(location[1]), 10, 0)
	if err != nil {
		return 0, false
	}
	atAddress, exists := source.Index(int(sourceOffset))
	if !exists {
		return 0, false
	}
	location = location[2:]
	if len(location) == 0 {
		return atAddress, true
	}
	switch location[0] {
	case 'l', 'r':
		leftMods, rightMods := c.getModifiers(atAddress)
		var mods []int
		if location[0] == 'l' {
			mods = leftMods
		} else {
			rightSlice := sort.IntSlice(rightMods)
			sort.Reverse(rightSlice)
			mods = []int(rightSlice)
		}
		if len(mods) == 0 {
			return 0, false
		}
		if len(location) > 1 && location[1] == '2' {
			if len(mods) > 1 {
				return mods[1], true
			}
		} else {
			return mods[0], true
		}
	case 'h':
		head, headExists := c.getHead(atAddress)
		if headExists {
			if len(location) > 1 && location[1] == '2' {
				headOfHead, headOfHeadExists := c.getHead(head.ID())
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

func (c *SimpleConfiguration) getModifierLabel(modifierID int) (string, bool) {
	arcs := c.Arcs().Get(&BasicDepArc{-1, "", modifierID})
	if len(arcs) > 0 {
		return string(arcs[0].GetRelation()), true
	}
	return "", false
}

func (c *SimpleConfiguration) Attribute(nodeID int, attribute []byte) (string, bool) {
	if nodeID < 0 || nodeID >= len(c.Nodes) {
		return "", false
	}
	switch attribute[0] {
	case 'd':
		return c.getConfDistance()
	case 'w':
		node := c.Nodes[nodeID]
		return node.Token, true
	case 'p':
		node := c.Nodes[nodeID]
		return node.POS, true
	case 'l':
		//		relation, relExists :=
		return c.getModifierLabel(nodeID)
	case 'v':
		if len(attribute) != 2 {
			return "", false
		}
		leftMods, rightMods := c.getModifiers(nodeID)
		switch attribute[1] {
		case 'l':
			return strconv.Itoa(len(leftMods)), true
		case 'r':
			return strconv.Itoa(len(rightMods)), true
		}
	case 's':
		if len(attribute) != 2 {
			return "", false
		}
		leftMods, rightMods := c.getModifiers(nodeID)
		var mods []int
		switch attribute[1] {
		case 'l':
			mods = leftMods
		case 'r':
			mods = rightMods
		}
		labels := make([]string, len(mods))
		for i, mod := range mods {
			labels[i], _ = c.getModifierLabel(mod)
		}
		return strings.Join(labels, SET_SEPARATOR), true
	}
	return "", false
}

func (c *SimpleConfiguration) getConfDistance() (string, bool) {
	stackTop, stackExists := c.Stack().Peek()
	queueTop, queueExists := c.Queue().Peek()
	if stackExists && queueExists {
		return strconv.Itoa(Util.AbsInt(queueTop - stackTop)), true
	}
	return "", false
}

func (c *SimpleConfiguration) getSource(location byte) Stack {
	switch location {
	case 'N':
		return c.Queue()
	case 'S':
		return c.Stack()
	}
	return nil
}

func (c *SimpleConfiguration) getHead(nodeID int) (*TaggedDepNode, bool) {
	arcs := c.Arcs().Get(&BasicDepArc{-1, "", nodeID})
	if len(arcs) == 0 {
		return nil, false
	}
	return c.Nodes[arcs[0].GetHead()], true
}

func (c *SimpleConfiguration) getModifiers(nodeID int) ([]int, []int) {
	arcs := c.Arcs().Get(&BasicDepArc{nodeID, "", -1})
	modifiers := make([]int, len(arcs))
	for i, arc := range arcs {
		modifiers[i] = arc.GetModifier()
	}
	sort.Ints(modifiers)
	var leftModifiers []int = modifiers[0:0]
	var rightModifiers []int = modifiers[0:0]
	for i, mod := range modifiers {
		if mod > nodeID {
			return leftModifiers[0:i], modifiers[i:len(modifiers)]
		}
	}
	return modifiers, rightModifiers
}
