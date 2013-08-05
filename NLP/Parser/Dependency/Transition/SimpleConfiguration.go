package Transition

import (
	"chukuparser/Algorithm/Graph"
	. "chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
	"chukuparser/Util"

	"fmt"
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
	Nodes    []TaggedDepNode
	previous *DependencyConfiguration
	Last     string
}

// Verify that SimpleConfiguration is a Configuration
var _ DependencyConfiguration = SimpleConfiguration{}

func (c SimpleConfiguration) Init(abstractSentence interface{}) {
	sent := abstractSentence.(TaggedSentence)
	// Nodes is always the same slice to the same token array
	c.Nodes = make([]TaggedDepNode, 0, len(sent)+1)
	c.Nodes = append(c.Nodes, TaggedDepNode{0, ROOT_TOKEN, ROOT_TOKEN})
	for i, taggedToken := range sent {
		c.Nodes = append(c.Nodes, TaggedDepNode{i + 1, taggedToken.Token, taggedToken.POS})
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
}

func (c SimpleConfiguration) Terminal() bool {
	return c.Queue().Size() == 0
}

func (c SimpleConfiguration) Stack() Stack {
	return c.stack
}

func (c SimpleConfiguration) Queue() Stack {
	return c.queue
}

func (c SimpleConfiguration) Arcs() ArcSet {
	return c.arcs
}

func (c SimpleConfiguration) Copy() *Configuration {
	newConf := new(SimpleConfiguration)

	newConf.stack = *(c.stack.Copy())
	newConf.queue = *(c.queue.Copy())
	newConf.arcs = *(c.arcs.Copy())

	newConf.Nodes = c.Nodes

	// store a pointer to the previous configuration
	previousConfig := DependencyConfiguration(c)
	newConf.previous = &previousConfig

	asConf := Configuration(*newConf)
	return &asConf
}

func (c SimpleConfiguration) Previous() *DependencyConfiguration {
	return c.previous
}

func (c SimpleConfiguration) SetLastTransition(t Transition) {
	(&c).Last = string(t)
}

func (c SimpleConfiguration) GetSequence() ConfigurationSequence {
	retval := make(ConfigurationSequence, 0, c.Arcs().Size())
	asDepConf := DependencyConfiguration(c)
	currentConf := &asDepConf
	for currentConf != nil {
		asConf := Configuration(*currentConf)
		retval = append(retval, &asConf)
		currentConf = (*currentConf).Previous()
	}
	return retval
}

// GRAPH FUNCTIONS
func (c SimpleConfiguration) GetVertices() []int {
	return Util.RangeInt(len(c.Nodes))
}

func (c SimpleConfiguration) GetEdges() []int {
	return Util.RangeInt(c.Arcs().Size())
}

func (c SimpleConfiguration) GetVertex(vertexID int) *Graph.Vertex {
	vertex := Graph.Vertex(c.Nodes[vertexID])
	return &vertex
}

func (c SimpleConfiguration) GetEdge(edgeID int) *Graph.Edge {
	arcPtr := c.Arcs().Index(edgeID)
	edge := Graph.Edge(*arcPtr)
	return &edge
}

func (c SimpleConfiguration) GetDirectedEdge(edgeID int) *Graph.DirectedEdge {
	arcPtr := c.Arcs().Index(edgeID)
	edge := Graph.DirectedEdge(*arcPtr)
	return &edge
}

func (c SimpleConfiguration) NumberOfVertices() int {
	return c.NumberOfNodes()
}

func (c SimpleConfiguration) NumberOfEdges() int {
	return c.NumberOfArcs()
}

func (c SimpleConfiguration) NumberOfNodes() int {
	return len(c.Nodes)
}

func (c SimpleConfiguration) NumberOfArcs() int {
	return c.Arcs().Size()
}

func (c SimpleConfiguration) GetNode(nodeID int) *NLP.DepNode {
	node := NLP.DepNode(c.Nodes[nodeID])
	return &node
}

func (c SimpleConfiguration) GetArc(arcID int) *NLP.DepArc {
	arcPtr := c.Arcs().Index(arcID)
	arc := NLP.DepArc(*arcPtr)
	return &arc
}

func (c SimpleConfiguration) GetLabeledArc(arcID int) *NLP.LabeledDepArc {
	arcPtr := c.Arcs().Index(arcID)
	arc := NLP.LabeledDepArc(*arcPtr)
	return &arc
}

// OUTPUT FUNCTIONS

func (c SimpleConfiguration) String() string {
	return fmt.Sprintf("%s\t=>([%s],\t[%s],\t[%s])",
		c.Last, c.StringStack(), c.StringQueue(),
		c.StringArcs())
}

func (c SimpleConfiguration) StringStack() string {
	switch {
	case c.Stack().Size() == 0:
		return ""
	case c.Stack().Size() <= 3:
		at0, _ := c.Nodes[c.Stack().Index(0)]
		at1, _ := c.Nodes[c.Stack().Index(1)]
		at2, _ := c.Nodes[c.Stack().Index(2)]
		return strings.Join([]string{at2, at1, at0}, ",")
	case c.Stack().Size() > 3:
		head, _ := c.Nodes[c.Stack().Index(0)]
		tail, _ := c.Nodes[c.Stack().Index(c.Stack().Size()-1)]
		return strings.Join([]string{tail, "...", head}, ",")
	}
	return ""
}

func (c SimpleConfiguration) StringQueue() string {
	switch {
	case c.Queue().Size() == 0:
		return ""
	case c.Queue().Size() <= 3:
		at0, _ := c.Nodes[c.Queue().Index(0)]
		at1, _ := c.Nodes[c.Queue().Index(1)]
		at2, _ := c.Nodes[c.Queue().Index(2)]
		return strings.Join([]string{at0, at1, at2}, ",")
	case c.Queue().Size() > 3:
		head, _ := c.Nodes[c.Queue().Index(0)]
		tail, _ := c.Nodes[c.Queue().Index(c.Queue().Size()-1)]
		return strings.Join([]string{head, "...", tail}, ",")
	}
}

func (c SimpleConfiguration) StringArcs() string {
	switch c.LastTrans[:2] {
	case "LA", "RA":
		lastArc := c.Arcs().Last()
		head := c.Nodes[lastArc.Head]
		mod := c.Nodes[lastArc.Modifier]
		arcStr := fmt.Sprintf("(%s,%s,%s)", head, lastArc.Relation, mod)
		return fmt.Sprintf("A%d=A%d+{%s}", c.Arcs.Size(), c.Arcs.Size()-1, arcStr)
	default:
		return fmt.Sprintf("A%d", c.Arcs().Size())
	}
}
