package Transition

import (
	"chukuparser/Algorithm/Graph"
	. "chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
	"chukuparser/Util"

	"fmt"
	"reflect"
	"strings"
)

type TaggedToken struct {
	Token string
	POS   string
}

type TaggedSentence []TaggedToken

const ROOT_TOKEN = "ROOT"

type SimpleConfiguration struct {
	stack    Stack
	queue    Stack
	arcs     ArcSet
	Nodes    []*TaggedDepNode
	previous DependencyConfiguration
	Last     string
}

// Verify that SimpleConfiguration is a Configuration
var _ DependencyConfiguration = &SimpleConfiguration{}

func (c *SimpleConfiguration) Init(abstractSentence interface{}) {
	sent := abstractSentence.(TaggedSentence)
	// Nodes is always the same slice to the same token array
	c.Nodes = make([]*TaggedDepNode, 0, len(sent)+1)
	c.Nodes = append(c.Nodes, &TaggedDepNode{0, ROOT_TOKEN, ROOT_TOKEN})
	for i, taggedToken := range sent {
		c.Nodes = append(c.Nodes, &TaggedDepNode{i + 1, taggedToken.Token, taggedToken.POS})
	}

	c.stack = NewStackArray(len(sent))
	c.queue = NewStackArray(len(sent))
	c.arcs = NewArcSetSimple(len(sent))

	// push index of ROOT node to Stack
	c.Stack().Push(0)
	// push indexes of statement nodes to Queue, in reverse order (first word at the top of the queue)
	for i := len(sent); i > 0; i-- {
		c.Queue().Push(i)
	}
	c.Last = ""
	c.previous = nil
}

func (c *SimpleConfiguration) Terminal() bool {
	return c.Queue().Size() == 0
}

func (c *SimpleConfiguration) Stack() Stack {
	return c.stack
}

func (c *SimpleConfiguration) Queue() Stack {
	return c.queue
}

func (c *SimpleConfiguration) Arcs() ArcSet {
	return c.arcs
}

func (c *SimpleConfiguration) Copy() Configuration {
	newConf := new(SimpleConfiguration)

	newConf.stack = c.Stack().Copy()
	newConf.queue = c.Queue().Copy()
	newConf.arcs = c.Arcs().Copy()

	newConf.Nodes = c.Nodes

	// store a pointer to the previous configuration
	newConf.previous = c
	newConf.Last = c.Last

	return newConf
}

func (c *SimpleConfiguration) Equal(other *SimpleConfiguration) bool {
	return c.Stack().Equal(other.Stack()) && c.Queue().Equal(other.Queue()) &&
		c.Arcs().Equal(other.Arcs()) && reflect.DeepEqual(c.Nodes, other.Nodes)
}

func (c *SimpleConfiguration) Previous() DependencyConfiguration {
	return c.previous
}

func (c *SimpleConfiguration) SetLastTransition(t Transition) {
	c.Last = string(t)
}

func (c *SimpleConfiguration) GetSequence() ConfigurationSequence {
	retval := make(ConfigurationSequence, 0, c.Arcs().Size())
	currentConf := DependencyConfiguration(c)
	for currentConf != nil {
		retval = append(retval, currentConf)
		currentConf = currentConf.Previous()
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
	return fmt.Sprintf("%s\t=>([%s],\t[%s],\t[%s])",
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
		for i := c.Queue().Size() - 1; i >= 0; i-- {
			atI, _ := c.Queue().Index(i)
			queueStrings = append(queueStrings, c.Nodes[atI].Token)
		}
		return strings.Join(queueStrings, ",")
	case queueSize > 3:
		headID, _ := c.Queue().Index(0)
		tailID, _ := c.Queue().Index(c.Queue().Size() - 1)
		head := c.Nodes[headID]
		tail := c.Nodes[tailID]
		return strings.Join([]string{tail.Token, "...", head.Token}, ",")
	default:
		return ""
	}
}

func (c *SimpleConfiguration) StringArcs() string {
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
