package Transition

import (
	AbstractTransition "chukuparser/Algorithm/Transition"
	"testing"
)

type SimpleConfTest struct {
	conf *SimpleConfiguration
	t    *testing.T
}

func (t *SimpleConfTest) Init() {
	c := t.conf
	sent := TaggedSentence{TaggedToken{"a", "NN"}, TaggedToken{"a", "NN"}}
	c.Init(sent)
	if c.Stack() == nil || c.Queue() == nil || c.Arcs() == nil {
		t.t.Error("Afte initialization got nil Stack/Queue/Arcs")
	}
	if len(c.Nodes) != 3 {
		t.t.Error("Got wrong size for Nodes slice")
	}
	if !(&c.Nodes[0]).Equal(&TaggedDepNode{0, ROOT_TOKEN, ROOT_TOKEN}) {
		t.t.Error("Init did not create root node")
	}
	if !(&c.Nodes[1]).Equal(&TaggedDepNode{1, sent[0].Token, sent[0].POS}) {
		t.t.Error("Init did not create node for tagged token")
	}
	if !(&c.Nodes[2]).Equal(&TaggedDepNode{2, sent[1].Token, sent[1].POS}) {
		t.t.Error("Init did not create node for tagged token")
	}
	if c.Stack().Size() != 1 {
		t.t.Error("Stack does not have correct initial size:")
	}
	sPeekVal, _ := c.Stack().Peek()
	if sPeekVal != 0 {
		t.t.Error("Stack head is not root node", sPeekVal)
	}
	if c.Queue().Size() != 2 {
		t.t.Error("Queue has wrong size")
	}
	qPeekVal, _ := c.Queue().Peek()
	if qPeekVal != 1 {
		t.t.Error("Queue head has wrong value")
	}
	qIdx1Val, _ := c.Queue().Index(1)
	if qIdx1Val != 2 {
		t.t.Error("Queue has wrong value at depth 1")
	}
	if c.Last != "" {
		t.t.Error("Wrong last action string")
	}
	if c.previous != nil {
		t.t.Error("Pointer to previous configuration is not nil")
	}
	if c.Arcs().Size() != 0 {
		t.t.Error("Initialized configuration has non-empty arc set")
	}
}

func (t *SimpleConfTest) Terminal() {
	c := t.conf
	c.Init(TaggedSentence{TaggedToken{"a", "NN"}})
	c.Queue().Clear()
	if !c.Terminal() {
		t.t.Error("Expected terminal configuration after queue cleared")
	}
	c.Queue().Push(0)
	if c.Terminal() {
		t.t.Error("Expected non-terminal configuration when queue is not empty")
	}
}

func (t *SimpleConfTest) Copy() {
	c := t.conf
	sent := TaggedSentence{TaggedToken{"a", "NN"}, TaggedToken{"a", "NN"}}
	c.Init(sent)
	newConf := c.Copy().(*SimpleConfiguration)
	if !c.Equal(newConf) {
		t.t.Error("Copy is not equal")
	}
	newConf.Stack().Push(5)
	if c.Equal(newConf) {
		t.t.Error("Copy is equal after stack push")
	}
	newConf.Stack().Pop()
	if !c.Equal(newConf) {
		t.t.Error("Copy is not equal after stack push,pop")
	}
	newConf.Queue().Push(0)
	if c.Equal(newConf) {
		t.t.Error("Copy is equal after queue push")
	}
	newConf.Queue().Pop()
	if !c.Equal(newConf) {
		t.t.Error("Copy is not equal after queue push,pop")
	}
	arc1, arc2 := &BasicDepArc{1, "a", 0}, &BasicDepArc{2, "b", 1}
	c.Arcs().Add(arc1)
	newConf.Arcs().Add(arc2)
	if c.Equal(newConf) {
		t.t.Error("Copy is equal after different arc set additions")
	}
	c.Arcs().Add(arc2)
	newConf.Arcs().Add(arc1)
	if !c.Equal(newConf) {
		t.t.Error("Copy is not equal after arc set additions in different order")
	}
}

func (t *SimpleConfTest) Address() {
}

func (t *SimpleConfTest) Arcs() {
	if t.conf.arcs != t.conf.Arcs() {
		t.t.Error("Returned wrong arcset object")
	}
}

func (t *SimpleConfTest) Attribute() {
}

func (t *SimpleConfTest) GetArc() {
}

func (t *SimpleConfTest) GetDirectedEdge() {
}

func (t *SimpleConfTest) GetEdge() {
}

func (t *SimpleConfTest) GetEdges() {
}

func (t *SimpleConfTest) GetLabeledArc() {
}

func (t *SimpleConfTest) GetNode() {
}

func (t *SimpleConfTest) GetSequence() {
}

func (t *SimpleConfTest) GetVertex() {
}

func (t *SimpleConfTest) GetVertices() {
}

func (t *SimpleConfTest) NumberOfArcs() {
	if t.conf.arcs.Size() != t.conf.NumberOfArcs() {
		t.t.Error("Reported wrong number of arcs")
	}
}

func (t *SimpleConfTest) NumberOfEdges() {
	if t.conf.arcs.Size() != t.conf.NumberOfEdges() {
		t.t.Error("Reported wrong number of edges")
	}
}

func (t *SimpleConfTest) NumberOfNodes() {
	if len(t.conf.Nodes) != t.conf.NumberOfNodes() {
		t.t.Error("Reported wrong number of nodes")
	}
}

func (t *SimpleConfTest) NumberOfVertices() {
	if len(t.conf.Nodes) != t.conf.NumberOfVertices() {
		t.t.Error("Reported wrong number of nodes")
	}
}

func (t *SimpleConfTest) Previous() {
	if t.conf.previous != t.conf.Previous() {
		t.t.Error("Reported wrong previous pointer")
	}
}

func (t *SimpleConfTest) Queue() {
	if t.conf.queue != t.conf.Queue() {
		t.t.Error("Returned wrong queue object")
	}
}

func (t *SimpleConfTest) SetLastTransition() {
	t.conf.SetLastTransition(AbstractTransition.Transition("bla"))
	if t.conf.Last != "bla" {
		t.t.Error("Setting last transition failed")
	}
}

func (t *SimpleConfTest) Stack() {
	if t.conf.stack != t.conf.Stack() {
		t.t.Error("Returned wrong stack object")
	}
}

func (t *SimpleConfTest) String() {
	str := t.conf.String()
	if len(str) == 0 {
		t.t.Error("Non empty configuration returns empty String")
	}
}

func (t *SimpleConfTest) StringArcs() {
	t.conf.SetLastTransition("LA")
	str := t.conf.StringArcs()
	if len(str) == 0 {
		t.t.Error("Non empty configuration returns empty StringArcs")
	}
}

func (t *SimpleConfTest) StringQueue() {
	str := t.conf.StringQueue()
	if len(str) == 0 {
		t.t.Error("Non empty configuration returns empty StringQueue")
	}
	t.conf.Queue().Clear()
	str = t.conf.StringQueue()
	if len(str) != 0 {
		t.t.Error("Empty queue in configuration returns non empty StringQueue")
	}
	t.conf.Queue().Push(0)
	t.conf.Queue().Push(0)
	t.conf.Queue().Push(0)
	t.conf.Queue().Push(0)
	str = t.conf.StringQueue()
	if len(str) == 0 {
		t.t.Error("Non-empty queue in configuration returns non empty StringQueue")
	}
}

func (t *SimpleConfTest) StringStack() {
	str := t.conf.StringStack()
	if len(str) == 0 {
		t.t.Error("Non empty configuration returns empty StringStack")
	}
	t.conf.Stack().Clear()
	str = t.conf.StringStack()
	if len(str) != 0 {
		t.t.Error("Empty stack in configuration returns non empty StringStack")
	}
	t.conf.Stack().Push(1)
	t.conf.Stack().Push(1)
	t.conf.Stack().Push(1)
	t.conf.Stack().Push(1)
	str = t.conf.StringStack()
	if len(str) == 0 {
		t.t.Error("Non-empty stack in configuration returns non empty StringStack")
	}
}

func (test *SimpleConfTest) All() {
	test.Init()
	test.Terminal()
	test.Copy()

	test.Address()
	test.Arcs()
	test.Attribute()
	test.GetArc()
	test.GetDirectedEdge()
	test.GetEdge()
	test.GetEdges()
	test.GetLabeledArc()
	test.GetNode()
	test.GetSequence()
	test.GetVertex()
	test.GetVertices()
	test.NumberOfArcs()
	test.NumberOfEdges()
	test.NumberOfNodes()
	test.NumberOfVertices()
	test.Previous()
	test.Queue()
	test.SetLastTransition()
	test.Stack()
	test.String()
	test.StringArcs()
	test.StringQueue()
	test.StringStack()
}

func TestSimpleConfiguration(t *testing.T) {
	test := SimpleConfTest{new(SimpleConfiguration), t}
	test.All()
}
