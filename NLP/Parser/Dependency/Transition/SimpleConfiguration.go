package Transition

import (
	. "chukuparser/Algorithm/Model/Perceptron"
	. "chukuparser/Algorithm/Transition"
	"fmt"
	// "math"
	// "regexp"
	// "strconv"
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
	Previous *Configuration
	Last     string
}

// Verify that SimpleConfiguration is a Configuration
var _ Configuration = SimpleConfiguration{}

func (c SimpleConfiguration) Stack() Stack {
	return c.stack
}

func (c SimpleConfiguration) Queue() Stack {
	return c.queue
}

func (c SimpleConfiguration) Arcs() ArcSet {
	return c.arcs
}

func (c SimpleConfiguration) Init(abstractSentence interface{}) {
	sent := abstractSentence.(TaggedSentence)
	// Nodes is always the same slice to the same token array
	c.Nodes = make([]DepNode, 0, sent.Size()+1)
	c.Nodes = append(c.Nodes, TaggedDepNode{ROOT_TOKEN, ROOT_TOKEN})
	for _, taggedToken := range sent {
		c.Nodes = append(c.Nodes, TaggedDepNode{taggedToken.Token, taggedToken.POS})
	}

	c.stack = NewStackArray(len(sent))
	c.queue = NewStackArray(len(sent))
	c.arcs = NewArcSetSimple()

	for i := len(sent) - 1; i >= 0; i-- {
		c.Queue().Push(i)
	}
	c.Last = ""
}

func (c SimpleConfiguration) Copy() *Configuration {
	newConf := new(SimpleConfiguration)

	newConf.stack = c.stack.Copy()
	newConf.queue = c.queue.Copy()
	newConf.arcs = c.arcs.Copy()

	newConf.Nodes = c.Nodes

	// store a pointer to the previous configuration
	newConf.Previous = c

	return newConf
}

func (c SimpleConfiguration) SetLastTransition(t Transition) {
	c.LastTrans = t
}

func (c SimpleConfiguration) Terminal() bool {
	return c.Queue().Size() == 0
}

func (c SimpleConfiguration) GetSequence() ConfigurationSequence {
	retval := make(ConfigurationSequence, 0, len(c.Arcs().Size()))
	currentConf := c
	for currentConf != nil {
		base = append(base, currentConf)
		currentConf = currentConf.Previous
	}
	return retval
}

func (c SimpleConfiguration) String() string {
	return fmt.Sprintf("%s\t=>([%s],\t[%s],\t[%s])",
		c.LastTrans, c.StringStack(), c.StringQueue(),
		c.StringArcs())
}

func (c SimpleConfiguration) StringStack() string {
	switch {
	case c.Stack.Size() == 0:
		return ""
	case c.Stack.Size() <= 3:
		at0, _ := c.Nodes[c.Stack.Index(0)]
		at1, _ := c.Nodes[c.Stack.Index(1)]
		at2, _ := c.Nodes[c.Stack.Index(2)]
		return strings.Join([...]string{at2, at1, at0}, ",")
	case c.Stack.Size() > 3:
		head, _ := c.Nodes[c.Stack.Index(0)]
		tail, _ := c.Nodes[c.Stack.Index(c.Stack.Size()-1)]
		return strings.Join([...]string{tail, "...", head}, ",")
	}
}

func (c SimpleConfiguration) StringQueue() string {
	switch {
	case c.Queue.Size() == 0:
		return ""
	case c.Queue.Size() <= 3:
		at0, _ := c.Nodes[c.Queue().Index(0)]
		at1, _ := c.Nodes[c.Queue().Index(1)]
		at2, _ := c.Nodes[c.Queue().Index(2)]
		return strings.Join([...]string{at0, at1, at2}, ",")
	case c.Queue.Size() > 3:
		head, _ := c.Nodes[c.Queue().Index(0)]
		tail, _ := c.Nodes[c.Queue().Index(c.Queue.Size()-1)]
		return strings.Join([...]string{head, "...", tail}, ",")
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
