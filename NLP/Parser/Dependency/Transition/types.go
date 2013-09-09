package Transition

import (
	"chukuparser/Algorithm/Graph"
	"chukuparser/Algorithm/Transition"
	NLP "chukuparser/NLP/Types"
	"chukuparser/Util"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"log"
)

type Stack interface {
	Clear()
	Push(int)
	Pop() (int, bool)
	Index(int) (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Stack
	Equal(Stack) bool
}

type Queue interface {
	Clear()
	Enqueue(int)
	Dequeue() (int, bool)
	Index(int) (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Queue
	Equal(Queue) bool
}

type ArcSet interface {
	Clear()
	Add(NLP.LabeledDepArc)
	Get(NLP.LabeledDepArc) []NLP.LabeledDepArc
	Size() int
	Last() NLP.LabeledDepArc
	Index(int) NLP.LabeledDepArc

	HasHead(int) bool
	HasModifiers(int) bool
	HasArc(int, int) bool

	Copy() ArcSet
	Equal(ArcSet) bool
}

type DependencyConfiguration interface {
	Util.Equaler
	Conf() Transition.Configuration
	Graph() NLP.LabeledDependencyGraph
	Address(location []byte, offset int) (int, bool)
	Attribute(source byte, nodeID int, attribute []byte) (interface{}, bool)
	Previous() DependencyConfiguration
	DecrementPointers()
	IncrementPointers()
	Clear()
}

type TaggedDepNode struct {
	Id       int
	Token    int
	POS      int
	TokenPOS int
	RawToken string
	RawPOS   string
}

var _ NLP.DepNode = &TaggedDepNode{}

func (t *TaggedDepNode) ID() int {
	return t.Id
}

func (t *TaggedDepNode) String() string {
	return t.RawToken
}

func (t *TaggedDepNode) Equal(otherEq Util.Equaler) bool {
	other := otherEq.(*TaggedDepNode)
	return reflect.DeepEqual(t, other)
}

type BasicDepArc struct {
	Head        int
	Relation    int
	Modifier    int
	RawRelation NLP.DepRel
}

var _ NLP.LabeledDepArc = &BasicDepArc{}

func (arc *BasicDepArc) ID() int {
	// a stand in for now
	return 0
}

func (arc *BasicDepArc) Vertices() []int {
	return []int{arc.Head, arc.Modifier}
}

func (arc *BasicDepArc) From() int {
	return arc.Modifier
}

func (arc *BasicDepArc) To() int {
	return arc.Head
}

func (arc *BasicDepArc) GetHead() int {
	return arc.Head
}

func (arc *BasicDepArc) GetModifier() int {
	return arc.Modifier
}

func (arc *BasicDepArc) GetRelation() NLP.DepRel {
	return arc.RawRelation
}

func (arc *BasicDepArc) Equal(otherEq Util.Equaler) bool {
	other := otherEq.(*BasicDepArc)
	return arc.Head == other.Head && arc.Modifier == other.Modifier && arc.RawRelation == other.RawRelation
}

func (arc *BasicDepArc) String() string {
	return fmt.Sprintf("(%d,%d-%s,%d)", arc.GetHead(), arc.Relation, arc.RawRelation, arc.GetModifier())
}

type BasicDepGraph struct {
	Nodes []NLP.DepNode
	Arcs  []*BasicDepArc
}

// Verify BasicDepGraph is a labeled dep. graph
var _ NLP.LabeledDependencyGraph = &BasicDepGraph{}

func (g *BasicDepGraph) GetVertices() []int {
	return Util.RangeInt(len(g.Nodes))
}

func (g *BasicDepGraph) GetEdges() []int {
	return Util.RangeInt(len(g.Arcs))
}

func (g *BasicDepGraph) GetVertex(n int) Graph.Vertex {
	if n >= len(g.Nodes) {
		return nil
	}
	return Graph.Vertex(g.Nodes[n])
}

func (g *BasicDepGraph) GetEdge(n int) Graph.Edge {
	if n >= len(g.Arcs) {
		return nil
	}
	return Graph.Edge(g.Arcs[n])
}

func (g *BasicDepGraph) GetDirectedEdge(n int) Graph.DirectedEdge {
	return Graph.DirectedEdge(g.Arcs[n])
}

func (g *BasicDepGraph) NumberOfVertices() int {
	return len(g.Nodes)
}

func (g *BasicDepGraph) NumberOfEdges() int {
	return len(g.Arcs)
}

func (g *BasicDepGraph) GetNode(n int) NLP.DepNode {
	if n >= len(g.Nodes) {
		return nil
	}
	return g.Nodes[n]
}

func (g *BasicDepGraph) GetArc(n int) NLP.DepArc {
	if n >= len(g.Arcs) {
		return nil
	}
	return NLP.DepArc(g.Arcs[n])
}

func (g *BasicDepGraph) NumberOfNodes() int {
	return g.NumberOfVertices()
}

func (g *BasicDepGraph) NumberOfArcs() int {
	return g.NumberOfEdges()
}

func (g *BasicDepGraph) GetLabeledArc(n int) NLP.LabeledDepArc {
	return NLP.LabeledDepArc(g.Arcs[n])
}

func (g *BasicDepGraph) StringEdges() string {
	arcs := make([]string, len(g.Arcs))
	for i, arc := range g.Arcs {
		arcs[i] = arc.String()
	}
	return strings.Join(arcs, "\n")
}

func (g *BasicDepGraph) Equal(otherEq Util.Equaler) bool {
	other := otherEq.(NLP.LabeledDependencyGraph)
	if g.NumberOfNodes() != other.NumberOfNodes() || g.NumberOfArcs() != other.NumberOfArcs() {
		return false
	}
	nodes, arcs := make([]NLP.DepNode, g.NumberOfNodes()), make([]NLP.LabeledDepArc, g.NumberOfArcs())
	for i := range other.GetVertices() {
		nodes[i] = other.GetNode(i)
	}
	for i := range other.GetEdges() {
		arcs[i] = other.GetLabeledArc(i)
	}
	nodesEqual := reflect.DeepEqual(g.Nodes, nodes)
	// numArcsNotEqual := 0
	otherArcSet := NewArcSetSimpleFromGraph(other)
	gArcSet := NewArcSetSimpleFromGraph(g)
	arcsEqual := gArcSet.Equal(otherArcSet)

	return nodesEqual && arcsEqual
}

func (g *BasicDepGraph) Sentence() NLP.Sentence {
	return NLP.Sentence(g.TaggedSentence())
}

func (g *BasicDepGraph) TaggedSentence() NLP.TaggedSentence {
	sent := make(NLP.BasicETaggedSentence, g.NumberOfNodes()-1)
	for _, node := range g.Nodes {
		taggedNode := node.(*TaggedDepNode)
		if taggedNode.RawToken == NLP.ROOT_TOKEN {
			continue
		}
		target := taggedNode.ID() - 1
		if target < 0 {
			panic("Too small")
		}
		if target >= len(sent) {
			panic("Too large")
		}
		sent[target] = NLP.EnumTaggedToken{
			NLP.TaggedToken{taggedNode.RawToken, taggedNode.RawPOS},
			taggedNode.Token,
			taggedNode.POS,
			taggedNode.TokenPOS,
		}
	}
	return NLP.TaggedSentence(NLP.EnumTaggedSentence(sent))
}

type ArcCachedDepNode struct {
	sync.RWMutex
	Node NLP.DepNode
	Head int
	// preallocated [3] array (most of the time <3 is needed)
	leftModArray, rightModArray [3]int
	leftMods, rightMods         []int
}

func (a *ArcCachedDepNode) LeftMods() []int {
	return a.leftMods
}

func (a *ArcCachedDepNode) RightMods() []int {
	return a.rightMods
}

func (a *ArcCachedDepNode) AddModifier(mod int) {
	// a.Lock()
	// defer a.Unlock()
	var (
		array *[3]int
		slice *[]int
	)
	if a.ID() > mod {
		array, slice = &a.leftModArray, &a.leftMods
	} else {
		array, slice = &a.rightModArray, &a.rightMods
	}
	if len(*slice) == cap(*array) {
		newslice := make([]int, len(*array), len(*array)+1)
		copy(newslice, *slice)
		slice = &newslice
	}

	// keep the slice sorted when adding
	var index, value int
	if len(*slice) == 0 || (*slice)[len(*slice)-1] < mod {
		*slice = append(*slice, mod)
	} else {
		log.Println("Add Modifier", *slice, mod)
		*slice = append(*slice, (*slice)[len(*slice)-1])
		log.Println("Appended", *slice)
		for i := len(*slice) - 2; i >= 0; i-- {
			log.Println("At (i,index)", i)
			value = (*slice)[i]
			(*slice)[index+1] = (*slice)[index]
			if value < mod {
				(*slice)[index] = mod
				break
			}
		}
		log.Println("Added", *slice)
	}
}

func NewArcCachedDepNode(from NLP.DepNode) *ArcCachedDepNode {
	a := &ArcCachedDepNode{
		Node: from,
		Head: -1,
	}
	a.leftMods, a.rightMods = a.leftModArray[0:0], a.rightModArray[0:0]
	return a
}

func (a *ArcCachedDepNode) ID() int {
	return a.Node.ID()
}

func (a *ArcCachedDepNode) String() string {
	return a.Node.String()
}

func (a *ArcCachedDepNode) Equal(otherEq Util.Equaler) bool {
	other := otherEq.(*ArcCachedDepNode)
	return reflect.DeepEqual(a.Node, other.Node) &&
		reflect.DeepEqual(a.leftMods, other.leftMods) &&
		reflect.DeepEqual(a.rightMods, other.rightMods)
}

func (a *ArcCachedDepNode) Copy() *ArcCachedDepNode {
	newNode := new(ArcCachedDepNode)
	*newNode = *a
	newNode.RWMutex = *new(sync.RWMutex)
	if len(a.leftMods) > cap(a.leftModArray) {
		newNode.leftMods = make([]int, len(a.leftMods))
		copy(newNode.leftMods, a.leftMods)
	} else {
		newNode.leftMods = newNode.leftModArray[:len(a.leftMods)]
	}
	if len(a.rightMods) > cap(a.rightModArray) {
		newNode.rightMods = make([]int, len(a.rightMods))
		copy(newNode.rightMods, a.rightMods)
	} else {
		newNode.rightMods = newNode.rightModArray[:len(a.rightMods)]
	}
	return newNode
}
