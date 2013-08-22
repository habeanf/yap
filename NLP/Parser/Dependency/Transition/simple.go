package Transition

import (
	"chukuparser/Algorithm/Graph"
	. "chukuparser/Algorithm/Transition"
	NLP "chukuparser/NLP/Types"
	"chukuparser/Util"

	"fmt"
	"reflect"
	"strings"
)

const ROOT_TOKEN = "ROOT"

type SimpleConfiguration struct {
	InternalStack    Stack
	InternalQueue    Stack
	InternalArcs     ArcSet
	Nodes            []*TaggedDepNode
	InternalPrevious *SimpleConfiguration
	Last             string
	Pointers         int
}

func (c *SimpleConfiguration) IncrementPointers() {
	c.Pointers++
}

func (c *SimpleConfiguration) DecrementPointers() {
	c.Pointers--
	if c.Pointers <= 0 {
		c.Clear()
	}
}

func (c *SimpleConfiguration) Conf() Configuration {
	return Configuration(c)
}

func (c *SimpleConfiguration) Graph() NLP.LabeledDependencyGraph {
	return NLP.LabeledDependencyGraph(c)
}

// Verify that SimpleConfiguration is a Configuration
var _ DependencyConfiguration = &SimpleConfiguration{}
var _ NLP.DependencyGraph = &SimpleConfiguration{}

func (c *SimpleConfiguration) ID() int {
	return 0
}

func (c *SimpleConfiguration) Init(abstractSentence interface{}) {
	sent := abstractSentence.(NLP.TaggedSentence)
	sentLength := len(sent.TaggedTokens())
	// Nodes is always the same slice to the same token array
	c.Nodes = make([]*TaggedDepNode, 1, sentLength+1)
	c.Nodes[0] = &TaggedDepNode{0, ROOT_TOKEN, ROOT_TOKEN}
	for i, taggedToken := range sent.TaggedTokens() {
		c.Nodes = append(c.Nodes, &TaggedDepNode{i + 1, taggedToken.Token, taggedToken.POS})
	}

	c.InternalStack = NewStackArray(sentLength)
	c.InternalQueue = NewStackArray(sentLength)
	c.InternalArcs = NewArcSetSimple(sentLength)

	// push index of ROOT node to Stack
	c.Stack().Push(0)
	// push indexes of statement nodes to Queue, in reverse order (first word at the top of the queue)
	for i := sentLength; i > 0; i-- {
		c.Queue().Push(i)
	}
	// explicit resetting of zero-valued properties
	// in case of reuse
	c.Last = ""
	c.InternalPrevious = nil
	c.Pointers = 0
}

func (c *SimpleConfiguration) Clear() {
	if c.Pointers > 0 {
		return
	}
	c.InternalStack = nil
	c.InternalQueue = nil
	c.InternalArcs = nil
	if c.InternalPrevious != nil {
		c.InternalPrevious.DecrementPointers()
		c.InternalPrevious.Clear()
		c.InternalPrevious = nil
	}
}

func (c *SimpleConfiguration) Terminal() bool {
	return c.Queue().Size() == 0
}

func (c *SimpleConfiguration) Stack() Stack {
	return c.InternalStack
}

func (c *SimpleConfiguration) Queue() Stack {
	return c.InternalQueue
}

func (c *SimpleConfiguration) Arcs() ArcSet {
	return c.InternalArcs
}

func (c *SimpleConfiguration) Copy() Configuration {
	newConf := new(SimpleConfiguration)

	if c.Stack() != nil {
		newConf.InternalStack = c.Stack().Copy()
	}
	if c.Queue() != nil {
		newConf.InternalQueue = c.Queue().Copy()
	}
	if c.Arcs() != nil {
		newConf.InternalArcs = c.Arcs().Copy()
	}
	newConf.Nodes = c.Nodes

	newConf.Last = c.Last

	// store a pointer to the previous configuration
	newConf.InternalPrevious = c
	// explicit setting of pointer counter
	newConf.Pointers = 0

	c.Pointers += 1

	return newConf
}

func (c *SimpleConfiguration) Equal(otherEq Util.Equaler) bool {
	switch other := otherEq.(type) {
	case *SimpleConfiguration:
		return c.Stack().Equal(other.Stack()) && c.Queue().Equal(other.Queue()) &&
			c.Arcs().Equal(other.Arcs()) && reflect.DeepEqual(c.Nodes, other.Nodes)
	case *BasicDepGraph:
		return other.Equal(c)
	}
	return false
}

func (c *SimpleConfiguration) Previous() DependencyConfiguration {
	return c.InternalPrevious
}

func (c *SimpleConfiguration) SetLastTransition(t Transition) {
	c.Last = string(t)
}

func (c *SimpleConfiguration) GetLastTransition() Transition {
	return Transition(c.Last)
}

func (c *SimpleConfiguration) GetSequence() ConfigurationSequence {
	if c.Arcs() == nil {
		return make(ConfigurationSequence, 0)
	}
	retval := make(ConfigurationSequence, 0, c.Arcs().Size())
	currentConf := c
	for currentConf != nil {
		retval = append(retval, currentConf)
		currentConf = currentConf.InternalPrevious
	}
	return retval
}

// GRAPH FUNCTIONS
func (c *SimpleConfiguration) GetVertices() []int {
	return Util.RangeInt(len(c.Nodes))
}

func (c *SimpleConfiguration) GetEdges() []int {
	return Util.RangeInt(c.Arcs().Size())
}

func (c *SimpleConfiguration) GetVertex(vertexID int) Graph.Vertex {
	return Graph.Vertex(c.Nodes[vertexID])
}

func (c *SimpleConfiguration) GetEdge(edgeID int) Graph.Edge {
	arcPtr := c.Arcs().Index(edgeID)
	return Graph.Edge(arcPtr)
}

func (c *SimpleConfiguration) GetDirectedEdge(edgeID int) Graph.DirectedEdge {
	arcPtr := c.Arcs().Index(edgeID)
	return Graph.DirectedEdge(arcPtr)
}

func (c *SimpleConfiguration) NumberOfVertices() int {
	return c.NumberOfNodes()
}

func (c *SimpleConfiguration) NumberOfEdges() int {
	return c.NumberOfArcs()
}

func (c *SimpleConfiguration) NumberOfNodes() int {
	return len(c.Nodes)
}

func (c *SimpleConfiguration) NumberOfArcs() int {
	return c.Arcs().Size()
}

func (c *SimpleConfiguration) GetNode(nodeID int) NLP.DepNode {
	// return NLP.DepNode(c.Nodes[nodeID])
	return NLP.DepNode(c.Nodes[nodeID])
}

func (c *SimpleConfiguration) GetArc(arcID int) NLP.DepArc {
	arcPtr := c.Arcs().Index(arcID)
	return NLP.DepArc(arcPtr)
}

func (c *SimpleConfiguration) GetLabeledArc(arcID int) NLP.LabeledDepArc {
	arcPtr := c.Arcs().Index(arcID)
	return NLP.LabeledDepArc(arcPtr)
}

// OUTPUT FUNCTIONS

func (c *SimpleConfiguration) String() string {
	return fmt.Sprintf("%s\t=>([%s],\t[%s],\t%s)",
		c.Last, c.StringStack(), c.StringQueue(),
		c.StringArcs())
}

func (c *SimpleConfiguration) StringStack() string {
	stackSize := c.Stack().Size()
	switch {
	case stackSize > 0 && stackSize <= 3:
		var stackStrings []string = make([]string, 0, 3)
		for i := c.Stack().Size() - 1; i >= 0; i-- {
			atI, _ := c.Stack().Index(i)
			stackStrings = append(stackStrings, c.Nodes[atI].Token)
		}
		return strings.Join(stackStrings, ",")
	case stackSize > 3:
		headID, _ := c.Stack().Index(0)
		tailID, _ := c.Stack().Index(c.Stack().Size() - 1)
		head := c.Nodes[headID]
		tail := c.Nodes[tailID]
		return strings.Join([]string{tail.Token, "...", head.Token}, ",")
	default:
		return ""
	}
}

func (c *SimpleConfiguration) StringQueue() string {
	queueSize := c.Queue().Size()
	switch {
	case queueSize > 0 && queueSize <= 3:
		var queueStrings []string = make([]string, 0, 3)
		for i := 0; i < c.Queue().Size(); i++ {
			atI, _ := c.Queue().Index(i)
			queueStrings = append(queueStrings, c.Nodes[atI].Token)
		}
		return strings.Join(queueStrings, ",")
	case queueSize > 3:
		headID, _ := c.Queue().Index(0)
		tailID, _ := c.Queue().Index(c.Queue().Size() - 1)
		head := c.Nodes[headID]
		tail := c.Nodes[tailID]
		return strings.Join([]string{head.Token, "...", tail.Token}, ",")
	default:
		return ""
	}
}

func (c *SimpleConfiguration) StringArcs() string {
	if len(c.Last) < 2 {
		return fmt.Sprintf("A%d", c.Arcs().Size())
	}
	switch c.Last[:2] {
	case "LA", "RA":
		lastArc := c.Arcs().Last()
		head := c.Nodes[lastArc.GetHead()]
		mod := c.Nodes[lastArc.GetModifier()]
		arcStr := fmt.Sprintf("(%s,%s,%s)", head.Token, string(lastArc.GetRelation()), mod.Token)
		return fmt.Sprintf("A%d=A%d+{%s}", c.Arcs().Size(), c.Arcs().Size()-1, arcStr)
	default:
		return fmt.Sprintf("A%d", c.Arcs().Size())
	}
}

func (c *SimpleConfiguration) StringGraph() string {
	return fmt.Sprintf("%v %v", c.Nodes, c.InternalArcs)
}

func (c *SimpleConfiguration) Sentence() NLP.Sentence {
	return NLP.Sentence(c.TaggedSentence())
}

func (c *SimpleConfiguration) TaggedSentence() NLP.TaggedSentence {
	sent := make([]NLP.TaggedToken, c.NumberOfNodes()-1)
	for i, taggedNode := range c.Nodes {
		if taggedNode.Token == ROOT_TOKEN {
			continue
		}
		sent[i] = NLP.TaggedToken{taggedNode.Token, taggedNode.POS}
	}
	return NLP.TaggedSentence(NLP.BasicTaggedSentence(sent))
}

func NewSimpleConfiguration() Configuration {
	return Configuration(new(SimpleConfiguration))
}
