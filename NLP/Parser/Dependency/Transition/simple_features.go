package Transition

import (
	NLP "chukuparser/NLP/Types"
	"chukuparser/Util"
	// "math"
	// "regexp"
	"sort"
	// "strconv"
	"strings"
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
		return 0, false
	}
	location = location[2:]
	if len(location) == 0 {
		return atAddress, true
	}
	switch location[0] {
	case 'l', 'r':
		leftMods, rightMods := c.GetModifiers(atAddress)
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
		node := c.Nodes[nodeID]
		return node.Token, true
	case 'p':
		node := c.Nodes[nodeID]
		// TODO: CPOS
		return node.POS, true
	case 'l':
		//		relation, relExists :=
		return c.GetModifierLabel(nodeID)
	case 'v':
		if len(attribute) != 2 {
			return 0, false
		}
		leftMods, rightMods := c.GetModifiers(nodeID)
		switch attribute[1] {
		case 'l':
			return len(leftMods), true
		case 'r':
			return len(rightMods), true
		}
	case 's':
		if len(attribute) != 2 {
			return 0, false
		}
		leftMods, rightMods := c.GetModifiers(nodeID)
		var mods []int
		switch attribute[1] {
		case 'l':
			mods = leftMods
		case 'r':
			mods = rightMods
		}
		labels := make([]string, len(mods))
		for i, mod := range mods {
			labelIndex, exists := c.GetModifierLabel(mod)
			if !exists {
				panic("Could not find label for modifier")
			}
			labels[i] = string(c.ERel.ValueOf(labelIndex).(NLP.DepRel))
		}
		return strings.Join(labels, SET_SEPARATOR), true
	}
	return 0, false
}

func (c *SimpleConfiguration) GetConfDistance() (int, bool) {
	stackTop, stackExists := c.Stack().Peek()
	queueTop, queueExists := c.Queue().Peek()
	if stackExists && queueExists {
		return Util.AbsInt(queueTop - stackTop), true
	}
	return 0, false
}

func (c *SimpleConfiguration) GetSource(location byte) Stack {
	switch location {
	case 'N':
		return c.Queue()
	case 'S':
		return c.Stack()
	}
	return nil
}

func (c *SimpleConfiguration) GetHead(nodeID int) (*TaggedDepNode, bool) {
	arcs := c.Arcs().Get(&BasicDepArc{-1, -1, nodeID, ""})
	if len(arcs) == 0 {
		return nil, false
	}
	return c.Nodes[arcs[0].GetHead()], true
}

func (c *SimpleConfiguration) GetModifiers(nodeID int) ([]int, []int) {
	arcs := c.Arcs().Get(&BasicDepArc{nodeID, -1, -1, ""})
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
