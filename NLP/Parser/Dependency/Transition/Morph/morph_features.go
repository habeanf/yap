package Morph

import (
	"chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
	"chukuparser/Util"
	"sort"
	"strconv"
	"strings"
)

func (m *MorphConfiguration) Attribute(source byte, nodeID int, attribute []byte) (interface{}, bool) {
	switch source {
	case 'A':
		arc := m.Arcs().Last()
		if arc == nil {
			return 0, false
		}
		head, mod := m.MorphNodes[arc.GetHead()], m.MorphNodes[arc.GetModifier()]
		switch attribute[0] {
		case 'g': // gen
			val, exists := mod.Features["gen"]
			other, otherExists := head.Features["gen"]
			if exists && otherExists && len(val) == len(other) {
				if val == other {
					return "1", true
				} else {
					return "0", true
				}
			}
			return 0, false
		case 'n': // num
			val, exists := mod.Features["num"]
			other, otherExists := head.Features["num"]
			if exists && otherExists {
				if val == "D" || other == "D" {
					return "1", true
				}
				if len(val) == len(other) {
					if val == other {
						return "1", true
					} else {
						return "0", true
					}
				}
			}
			return 0, false
		case 'p': // per
			val, exists := mod.Features["per"]
			other, otherExists := head.Features["per"]
			if exists && otherExists {
				if val == "A" || other == "A" {
					return "1", true
				}
				if val == other {
					return "1", true
				} else {
					return "0", true
				}
			}
			return 0, false
		case 'o': // polar
			val, exists := mod.Features["polar"]
			other, otherExists := head.Features["polar"]
			if exists && otherExists {
				if val == other {
					return "1", true
				} else {
					return "0", true
				}
			}
			return 0, false
		case 't': // tense
			val, exists := mod.Features["tense"]
			other, otherExists := head.Features["tense"]
			if exists && otherExists {
				if val == other {
					return "1", true
				} else {
					return "0", true
				}
			}
			return 0, false
		default:
			panic("Unknown attribute " + string(attribute))
		}
	case 'M':
		latId, exists := m.LatticeQueue.Index(nodeID)
		if !exists {
			return 0, false
		}
		switch attribute[0] {
		case 'w':
			lat := m.Lattices[latId]
			return string(lat.Token), true
		default:
			panic("Unknown attribute " + string(attribute))
		}
	case 'N', 'S':
		if nodeID < 0 || nodeID >= len(m.MorphNodes) {
			return 0, false
		}
		switch attribute[0] {
		case 't':
			return m.GetQueueMorphs()
		case 'd':
			return m.GetConfDistance()
		case 'w':
			node := m.MorphNodes[nodeID]
			if len(attribute) == 2 && attribute[1] == 'p' {
				return node.EFCPOS, true
			}
			return node.EForm, true
		case 'p':
			node := m.MorphNodes[nodeID]
			return node.EPOS, true
		case 'l':
			//		relation, relExists :=
			return m.GetModifierLabel(nodeID)
		case 'v':
			if len(attribute) != 2 {
				return 0, false
			}
			leftMods, rightMods := m.GetNumModifiers(nodeID)
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
		default:
			panic("Unknown attribute " + string(attribute))
		}
	default:
		panic("Unknown attribute " + string(attribute))
	}
	return 0, false
}

func (m *MorphConfiguration) GetHead(nodeID int) (*NLP.EMorpheme, bool) {
	arcs := m.Arcs().Get(&Transition.BasicDepArc{-1, -1, nodeID, ""})
	if len(arcs) == 0 {
		return nil, false
	}
	return m.MorphNodes[arcs[0].GetHead()], true
}

func (m *MorphConfiguration) GetModifiers(nodeID int) ([]int, []int) {
	arcs := m.Arcs().Get(&Transition.BasicDepArc{nodeID, -1, -1, ""})
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

func (m *MorphConfiguration) GetNumModifiers(nodeID int) (int, int) {
	arcs := m.Arcs().Get(&Transition.BasicDepArc{nodeID, -1, -1, ""})
	var left, right int
	for _, arc := range arcs {
		if arc.GetModifier() > nodeID {
			left++
		} else {
			right++
		}
	}
	return left, right
}

func (m *MorphConfiguration) GetSource(location byte) interface{} {
	switch location {
	case 'N':
		return m.Queue()
	case 'S':
		return m.Stack()
	case 'A':
		return m.Arcs()
	case 'M':
		return m.LatticeQueue
	default:
		panic("Unknown location " + string(location))
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
	arcs := m.Arcs().Get(&Transition.BasicDepArc{-1, -1, modifierID, ""})
	if len(arcs) > 0 {
		return string(arcs[0].GetRelation()), true
	}
	return "", false
}

func (m *MorphConfiguration) Address(location []byte, sourceOffset int) (int, bool) {
	s := m.GetSource(location[0])
	if s == nil {
		return 0, false
	}
	// sourceOffset, err := strconv.ParseInt(string(location[1]), 10, 0)
	// if err != nil {
	// 	return 0, false
	// }
	location = location[2:]
	switch source := s.(type) {
	case *Transition.StackArray:
		atAddress, exists := source.Index(int(sourceOffset))
		if !exists {
			return 0, false
		}
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
		default:
			panic("Unknown location " + string(location))
		}
	case *Transition.ArcSetSimple:
		return 0, true
	default:
		panic("Unknown location " + string(location))
	}
	return 0, false
}
