package Morph

import (
	G "chukuparser/Algorithm/Graph"
	"chukuparser/Algorithm/Transition"
	. "chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
	"chukuparser/Util"
	"fmt"
	// "log"
	"reflect"
	"strings"
)

type MorphConfiguration struct {
	SimpleConfiguration
	LatticeQueue Stack
	Lattices     []NLP.Lattice
	Mappings     []*NLP.Mapping
	// MorphNodes    []*NLP.EMorpheme
	MorphPrevious *MorphConfiguration
}

// Verify that MorphConfiguration is a Configuration
var _ DependencyConfiguration = &MorphConfiguration{}
var _ NLP.MorphDependencyGraph = &MorphConfiguration{}

func (m *MorphConfiguration) Init(abstractLattice interface{}) {
	// note: doesn't call SimpleConfiguration's init
	// because we don't want to initialize the "Nodes" variable in
	// the struct
	latticeSent := abstractLattice.(NLP.LatticeSentence)
	sentLength := len(latticeSent)

	m.Lattices = make(NLP.LatticeSentence, sentLength+1)
	m.Lattices[0] = NLP.NewRootLattice()
	copy(m.Lattices[1:], latticeSent)

	maxSentLength := 0
	var latP *NLP.Lattice
	for _, lat := range m.Lattices {
		latP = &lat
		maxSentLength += latP.MaxPathLen()
	}

	// regular configuration
	m.InternalStack = NewStackArray(maxSentLength)
	m.InternalQueue = NewStackArray(maxSentLength)
	m.InternalArcs = NewArcSetSimple(maxSentLength)

	m.LatticeQueue = NewStackArray(sentLength)
	// m.MorphNodes = make([]*NLP.EMorpheme, 1, maxSentLength)

	// m.MorphNodes[0] = &NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{0, 0, 0}, "ROOT", "ROOT", "ROOT", nil, 0}}

	m.Nodes = make([]*ArcCachedDepNode, 1, maxSentLength)
	m.Nodes[0] = NewArcCachedDepNode(NLP.DepNode(&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{0, 0, 0}, "ROOT", "ROOT", "ROOT", nil, 0}}))
	m.Mappings = make([]*NLP.Mapping, 1, len(m.Lattices))
	m.Mappings[0] = &NLP.Mapping{"ROOT", []*NLP.EMorpheme{m.GetMorpheme(0)}}

	// push index of ROOT node to Stack
	m.Stack().Push(0)

	// push indexes of statement nodes to *LatticeQueue*, in reverse order (first word at the top of the queue)
	for i := sentLength; i > 0; i-- {
		m.LatticeQueue.Push(i)
	}

	// explicit resetting of zero-valued properties
	// in case of reuse
	m.Last = -1
	m.InternalPrevious = nil
	m.MorphPrevious = nil
	m.Pointers = 0
}

func (m *MorphConfiguration) Copy() Transition.Configuration {
	newConf := new(MorphConfiguration)
	newSimple := m.SimpleConfiguration.Copy().(*SimpleConfiguration)
	newConf.SimpleConfiguration = *newSimple

	newConf.Mappings = make([]*NLP.Mapping, len(m.Mappings), len(m.Lattices))
	copy(newConf.Mappings, m.Mappings)

	if m.LatticeQueue != nil {
		newConf.LatticeQueue = m.LatticeQueue.Copy()
	}
	newConf.EWord, newConf.EPOS, newConf.EWPOS, newConf.ERel, newConf.ETrans = m.EWord, m.EPOS, m.EWPOS, m.ERel, m.ETrans

	// lattices slice is read only, no need for copy
	newConf.Lattices = m.Lattices
	newConf.MorphPrevious = m
	return newConf
}

func (m *MorphConfiguration) Equal(otherEq Util.Equaler) bool {
	switch other := otherEq.(type) {
	case *MorphConfiguration:
		// log.Println("Equality")
		// log.Println("\n", m.GetSequence())
		// log.Println("\n", other.GetSequence())
		if !((&m.SimpleConfiguration).Equal(&other.SimpleConfiguration)) {
			return false
		}
		return m.NumberOfNodes() == other.NumberOfNodes() &&
			m.NumberOfArcs() == other.NumberOfArcs() &&
			reflect.DeepEqual(m.Lattices, other.Lattices) &&
			reflect.DeepEqual(m.Mappings, other.Mappings) &&
			reflect.DeepEqual(m.Nodes, other.Nodes) &&
			m.LatticeQueue.Equal(other.LatticeQueue)

	case *BasicDepGraph:
		return other.Equal(m)
	}
	return false
}

func (m *MorphConfiguration) Graph() NLP.LabeledDependencyGraph {
	return NLP.LabeledDependencyGraph(m)
}

func (m *MorphConfiguration) Terminal() bool {
	return m.LatticeQueue.Size() == 0 && m.SimpleConfiguration.Queue().Size() == 0
}

func (m *MorphConfiguration) GetMappings() []*NLP.Mapping {
	return m.Mappings
}

func (m *MorphConfiguration) GetMorpheme(i int) *NLP.EMorpheme {
	return m.Nodes[i].Node.(*NLP.EMorpheme)
}

// OUTPUT FUNCTIONS
// TODO: fix this
func (m *MorphConfiguration) String() string {
	var trans string
	if m.Last < 0 {
		trans = ""
	} else {
		trans = m.ETrans.ValueOf(int(m.Last)).(string)
	}
	return fmt.Sprintf("%s\t=>\t([%s],\t[%s],\t[%s],\t%s, \t%s)",
		trans, m.StringStack(), m.StringQueue(),
		m.StringLatticeQueue(),
		m.StringArcs(),
		m.StringMappings())
}

func (m *MorphConfiguration) StringLatticeQueue() string {
	queueSize := m.LatticeQueue.Size()
	switch {
	case queueSize > 0 && queueSize <= 3:
		var queueStrings []string = make([]string, 0, 3)
		for i := 0; i < m.LatticeQueue.Size(); i++ {
			atI, _ := m.LatticeQueue.Index(i)
			queueStrings = append(queueStrings, string(m.Lattices[atI].Token))
		}
		return strings.Join(queueStrings, ",")
	case queueSize > 3:
		headID, _ := m.LatticeQueue.Index(0)
		tailID, _ := m.LatticeQueue.Index(m.LatticeQueue.Size() - 1)
		head := m.Lattices[headID]
		tail := m.Lattices[tailID]
		return strings.Join([]string{string(head.Token), "...", string(tail.Token)}, ",")
	default:
		return ""
	}

}

func (m *MorphConfiguration) StringStack() string {
	stackSize := m.Stack().Size()
	switch {
	case stackSize > 0 && stackSize <= 3:
		var stackStrings []string = make([]string, 0, 3)
		for i := m.Stack().Size() - 1; i >= 0; i-- {
			atI, _ := m.Stack().Index(i)
			stackStrings = append(stackStrings, m.GetMorpheme(atI).Form)
		}
		return strings.Join(stackStrings, ",")
	case stackSize > 3:
		headID, _ := m.Stack().Index(0)
		tailID, _ := m.Stack().Index(m.Stack().Size() - 1)
		head := m.GetMorpheme(headID)
		tail := m.GetMorpheme(tailID)
		return strings.Join([]string{tail.Form, "...", head.Form}, ",")
	default:
		return ""
	}
}

func (m *MorphConfiguration) StringArcs() string {
	if m.Last < 0 {
		return "A0"
	}
	last := m.ETrans.ValueOf(int(m.Last)).(string)
	if len(last) < 2 {
		return fmt.Sprintf("A%d", m.Arcs().Size())
	}
	switch last[:2] {
	case "LA", "RA":
		lastArc := m.Arcs().Last()
		head := m.GetMorpheme(lastArc.GetHead())
		mod := m.GetMorpheme(lastArc.GetModifier())
		arcStr := fmt.Sprintf("(%s,%s,%s)", head.Form, lastArc.GetRelation().String(), mod.Form)
		return fmt.Sprintf("A%d=A%d+{%s}", m.Arcs().Size(), m.Arcs().Size()-1, arcStr)
	default:
		return fmt.Sprintf("A%d", m.Arcs().Size())
	}
}

func (m *MorphConfiguration) StringMappings() string {
	if m.Last < 0 {
		return "M0"
	}
	mappingLen := len(m.Mappings) - 1
	last := m.ETrans.ValueOf(int(m.Last)).(string)
	if len(last) < 2 || last[:2] == "MD" {
		lastMap := m.Mappings[mappingLen]
		mapStr := fmt.Sprintf("(%s,%s)", lastMap.Token, lastMap.Spellout.AsString())
		if mappingLen == 0 {
			return fmt.Sprintf("M%d={%s}", mappingLen, mapStr)
		} else {
			return fmt.Sprintf("M%d=M%d+{%s}", mappingLen, mappingLen-1, mapStr)
		}
	} else {
		return fmt.Sprintf("M%d", mappingLen)
	}
}

func (m *MorphConfiguration) StringQueue() string {
	queueSize := m.Queue().Size()
	switch {
	case queueSize > 0 && queueSize <= 3:
		var queueStrings []string = make([]string, 0, 3)
		for i := 0; i < m.Queue().Size(); i++ {
			atI, _ := m.Queue().Index(i)
			queueStrings = append(queueStrings, m.GetMorpheme(atI).Form)
		}
		return strings.Join(queueStrings, ",")
	case queueSize > 3:
		headID, _ := m.Queue().Index(0)
		tailID, _ := m.Queue().Index(m.Queue().Size() - 1)
		head := m.GetMorpheme(headID)
		tail := m.GetMorpheme(tailID)
		return strings.Join([]string{head.Form, "...", tail.Form}, ",")
	default:
		return ""
	}
}

func (m *MorphConfiguration) Conf() Transition.Configuration {
	return Transition.Configuration(m)
}

func (m *MorphConfiguration) Previous() DependencyConfiguration {
	return m.MorphPrevious
}

func (m *MorphConfiguration) GetSequence() Transition.ConfigurationSequence {
	if m.Arcs() == nil {
		return make(Transition.ConfigurationSequence, 0)
	}
	retval := make(Transition.ConfigurationSequence, 0, m.Arcs().Size())
	currentConf := m
	for currentConf != nil {
		retval = append(retval, currentConf)
		currentConf = currentConf.MorphPrevious
	}
	return retval
}

func (m *MorphConfiguration) GetVertices() []int {
	return Util.RangeInt(len(m.Nodes))
}

func (m *MorphConfiguration) GetNode(nodeID int) NLP.DepNode {
	return NLP.DepNode(m.Nodes[nodeID])
}

func NewMorphConfiguration() Transition.Configuration {
	return Transition.Configuration(new(MorphConfiguration))
}
