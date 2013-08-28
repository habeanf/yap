package Morph

import (
	"chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
	"chukuparser/Util"
	"sort"
	"strconv"
	"strings"
)

func (m *MorphConfiguration) Attribute(nodeID int, attribute []byte) (string, bool) {
	if nodeID < 0 || nodeID >= len(m.MorphNodes) {
		return "", false
	}
	switch attribute[0] {
	case 't':
		return m.GetQueueMorphs()
	case 'd':
		return m.GetConfDistance()
	case 'w':
		node := m.MorphNodes[nodeID]
		return node.Form, true
	case 'p':
		node := m.MorphNodes[nodeID]
		return node.POS, true
	case 'l':
		//		relation, relExists :=
		return m.GetModifierLabel(nodeID)
	case 'v':
		if len(attribute) != 2 {
			return "", false
		}
		leftMods, rightMods := m.GetModifiers(nodeID)
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
		leftMods, rightMods := m.GetModifiers(nodeID)
		var mods []int
		switch attribute[1] {
		case 'l':
			mods = leftMods
		case 'r':
			mods = rightMods
		}
		labels := make([]string, len(mods))
		for i, mod := range mods {
			labels[i], _ = m.GetModifierLabel(mod)
		}
		return strings.Join(labels, Transition.SET_SEPARATOR), true
	}
	return "", false
}

func (m *MorphConfiguration) GetHead(nodeID int) (*NLP.Morpheme, bool) {
	arcs := m.Arcs().Get(&Transition.BasicDepArc{-1, "", nodeID})
	if len(arcs) == 0 {
		return nil, false
	}
	return m.MorphNodes[arcs[0].GetHead()], true
}

func (m *MorphConfiguration) GetModifiers(nodeID int) ([]int, []int) {
	arcs := m.Arcs().Get(&Transition.BasicDepArc{nodeID, "", -1})
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

func (m *MorphConfiguration) GetSource(location byte) Transition.Stack {
	switch location {
	case 'N':
		return m.Queue()
	case 'S':
		return m.Stack()
	}
	return nil
}

func (m *MorphConfiguration) GetQueueMorphs() (string, bool) {
	if m.Queue().Size() == 0 {
		return "", false
	}
	strs := make([]string, m.Queue().Size())
	for i := 0; i < m.Queue().Size(); i++ {
		atI := m.MorphNodes[i]
		strs[i] = atI.CPOS
	}
	return strings.Join(strs, "-"), true
}

func (m *MorphConfiguration) GetConfDistance() (string, bool) {
	stackTop, stackExists := m.Stack().Peek()
	queueTop, queueExists := m.Queue().Peek()
	if stackExists && queueExists {
		return strconv.Itoa(Util.AbsInt(queueTop - stackTop)), true
	}
	return "", false
}

func (m *MorphConfiguration) GetModifierLabel(modifierID int) (string, bool) {
	arcs := m.Arcs().Get(&Transition.BasicDepArc{-1, "", modifierID})
	if len(arcs) > 0 {
		return string(arcs[0].GetRelation()), true
	}
	return "", false
}

func (m *MorphConfiguration) Address(location []byte) (int, bool) {
	source := m.GetSource(location[0])
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
		leftMods, rightMods := m.GetModifiers(atAddress)
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
		head, headExists := m.GetHead(atAddress)
		if headExists {
			if len(location) > 1 && location[1] == '2' {
				headOfHead, headOfHeadExists := m.GetHead(head.ID())
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
