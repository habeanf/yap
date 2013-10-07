package transition

import (
	"chukuparser/algorithm/graph"
	"chukuparser/algorithm/transition"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"
	"fmt"
	// "log"
	"reflect"
	"strings"
)

type Index interface {
	Index(int) (int, bool)
}

type Stack interface {
	Index
	Clear()
	Push(int)
	Pop() (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Stack
	Equal(Stack) bool
}

type Queue interface {
	Index
	Clear()
	Enqueue(int)
	Dequeue() (int, bool)
	Pop() (int, bool)
	Peek() (int, bool)
	Size() int

	Copy() Queue
	Equal(Queue) bool
}

type ArcSet interface {
	Clear()
	Add(nlp.LabeledDepArc)
	Get(nlp.LabeledDepArc) []nlp.LabeledDepArc
	Size() int
	Last() nlp.LabeledDepArc
	Index(int) nlp.LabeledDepArc

	HasHead(int) bool
	HasModifiers(int) bool
	HasArc(int, int) bool

	Copy() ArcSet
	Equal(ArcSet) bool
}

type DependencyConfiguration interface {
	util.Equaler
	Conf() transition.Configuration
	Graph() nlp.LabeledDependencyGraph
	Address(location []byte, offset int) (int, bool)
	Attribute(source byte, nodeID int, attribute []byte) (interface{}, bool)
	Previous() DependencyConfiguration
	DecrementPointers()
	IncrementPointers()
	Clear()
	GetLastTransition() transition.Transition
	Copy() transition.Configuration
}

type TaggedDepNode struct {
	Id       int
	Token    int
	POS      int
	TokenPOS int
	RawToken string
	RawPOS   string
}

var _ nlp.DepNode = &TaggedDepNode{}

func (t *TaggedDepNode) ID() int {
	return t.Id
}

func (t *TaggedDepNode) String() string {
	return t.RawToken
}

func (t *TaggedDepNode) Equal(otherEq util.Equaler) bool {
	other := otherEq.(*TaggedDepNode)
	return reflect.DeepEqual(t, other)
}

type BasicDepArc struct {
	Head        int
	Relation    int
	Modifier    int
	RawRelation nlp.DepRel
}

var _ nlp.LabeledDepArc = &BasicDepArc{}

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

func (arc *BasicDepArc) GetRelation() nlp.DepRel {
	return arc.RawRelation
}

func (arc *BasicDepArc) Equal(otherEq util.Equaler) bool {
	other := otherEq.(*BasicDepArc)
	return arc.Head == other.Head && arc.Modifier == other.Modifier && arc.RawRelation == other.RawRelation
}

func (arc *BasicDepArc) String() string {
	return fmt.Sprintf("(%d,%d-%s,%d)", arc.GetHead(), arc.Relation, arc.RawRelation, arc.GetModifier())
}

type BasicDepGraph struct {
	Nodes []nlp.DepNode
	Arcs  []*BasicDepArc
}

// Verify BasicDepGraph is a labeled dep. graph
var _ nlp.LabeledDependencyGraph = &BasicDepGraph{}

func (g *BasicDepGraph) GetVertices() []int {
	return util.RangeInt(len(g.Nodes))
}

func (g *BasicDepGraph) GetEdges() []int {
	return util.RangeInt(len(g.Arcs))
}

func (g *BasicDepGraph) GetVertex(n int) graph.Vertex {
	if n >= len(g.Nodes) {
		return nil
	}
	return graph.Vertex(g.Nodes[n])
}

func (g *BasicDepGraph) GetEdge(n int) graph.Edge {
	if n >= len(g.Arcs) {
		return nil
	}
	return graph.Edge(g.Arcs[n])
}

func (g *BasicDepGraph) GetDirectedEdge(n int) graph.DirectedEdge {
	return graph.DirectedEdge(g.Arcs[n])
}

func (g *BasicDepGraph) NumberOfVertices() int {
	return len(g.Nodes)
}

func (g *BasicDepGraph) NumberOfEdges() int {
	return len(g.Arcs)
}

func (g *BasicDepGraph) GetNode(n int) nlp.DepNode {
	if n >= len(g.Nodes) {
		return nil
	}
	return g.Nodes[n]
}

func (g *BasicDepGraph) GetArc(n int) nlp.DepArc {
	if n >= len(g.Arcs) {
		return nil
	}
	return nlp.DepArc(g.Arcs[n])
}

func (g *BasicDepGraph) NumberOfNodes() int {
	return g.NumberOfVertices()
}

func (g *BasicDepGraph) NumberOfArcs() int {
	return g.NumberOfEdges()
}

func (g *BasicDepGraph) GetLabeledArc(n int) nlp.LabeledDepArc {
	return nlp.LabeledDepArc(g.Arcs[n])
}

func (g *BasicDepGraph) StringEdges() string {
	arcs := make([]string, len(g.Arcs))
	for i, arc := range g.Arcs {
		arcs[i] = arc.String()
	}
	return strings.Join(arcs, "\n")
}

func (g *BasicDepGraph) Equal(otherEq util.Equaler) bool {
	other := otherEq.(nlp.LabeledDependencyGraph)
	if g.NumberOfNodes() != other.NumberOfNodes() || g.NumberOfArcs() != other.NumberOfArcs() {
		return false
	}
	nodes, arcs := make([]nlp.DepNode, g.NumberOfNodes()), make([]nlp.LabeledDepArc, g.NumberOfArcs())
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

func (g *BasicDepGraph) Sentence() nlp.Sentence {
	return nlp.Sentence(g.TaggedSentence())
}

func (g *BasicDepGraph) TaggedSentence() nlp.TaggedSentence {
	sent := make(nlp.BasicETaggedSentence, g.NumberOfNodes())
	for _, node := range g.Nodes {
		taggedNode := node.(*TaggedDepNode)
		target := taggedNode.ID()
		if target < 0 {
			panic(fmt.Sprintf("Too small: %d ", target))
		}
		if target >= len(sent) {
			panic(fmt.Sprintf("Too large; Size is %d got target %d", len(sent), target))
		}
		sent[target] = nlp.EnumTaggedToken{
			nlp.TaggedToken{taggedNode.RawToken, taggedNode.RawPOS},
			taggedNode.Token,
			taggedNode.POS,
			taggedNode.TokenPOS,
		}
	}
	return nlp.TaggedSentence(nlp.EnumTaggedSentence(sent))
}

type ArcCachedDepNode struct {
	Node         nlp.DepNode
	Head, ELabel int
	// preallocated [3] arrays (most of the time <3 is needed)
	leftModArray, rightModArray     [3]int
	leftMods, rightMods             []int
	leftLabelArray, rightLabelArray [3]int
	leftLabels, rightLabels         []int
}

func (a *ArcCachedDepNode) LeftMods() []int {
	return a.leftMods
}

func (a *ArcCachedDepNode) RightMods() []int {
	return a.rightMods
}

func (a *ArcCachedDepNode) LeftLabelSet() interface{} {
	return GetArrayInt(a.leftLabels)
}

func (a *ArcCachedDepNode) RightLabelSet() interface{} {
	return GetArrayInt(a.rightLabels)
}

func (a *ArcCachedDepNode) LRSortedInsertion(slice *[]int, val int) {
	newslice := *slice
	if len(newslice) > 0 {
		for _, cur := range newslice {
			if cur == val {
				return
			}
		}
	}
	if len(*slice) == cap(*slice) {
		newslice = make([]int, len(*slice), len(*slice)+1)
		copy(newslice, *slice)
	}

	// keep the slice sorted when adding
	var value int
	if len(newslice) == 0 || (newslice)[len(newslice)-1] < val {
		newslice = append(newslice, val)
	} else {
		newslice = append(newslice, (newslice)[len(newslice)-1])
		for i := len(newslice) - 2; i >= 0; i-- {
			value = (newslice)[i]
			// log.Println("At i", i, "value", value, "copying to", i+1)
			(newslice)[i+1] = (newslice)[i]
			// log.Println(newslice)
			if value == val {
				// log.Println("Breaking (1)")
				return
			}
			if value < val {
				// log.Println("Breaking (2)")
				(newslice)[i] = val
				break
			}
			if i == 0 {
				(newslice)[i] = val
				// log.Println(newslice)
			}
		}
	}
	*slice = newslice
}

func (a *ArcCachedDepNode) AddModifier(mod int, label int) {
	if a.ID() > mod {
		// log.Println("Adding mod", mod)
		a.LRSortedInsertion(&a.leftMods, mod)
		// log.Println("Adding label", label)
		// log.Println("Left labels before:", a.leftLabels)
		a.LRSortedInsertion(&a.leftLabels, label)
		// log.Println("Left labels after:", a.leftLabels)
	} else {
		// log.Println("Adding mod", mod)
		a.LRSortedInsertion(&a.rightMods, mod)
		// log.Println("Adding label", label)
		// log.Println("Right labels before:", a.leftLabels)
		a.LRSortedInsertion(&a.rightLabels, label)
		// log.Println("Right labels after:", a.leftLabels)
	}
}

func NewArcCachedDepNode(from nlp.DepNode) *ArcCachedDepNode {
	a := &ArcCachedDepNode{
		Node:   from,
		Head:   -1,
		ELabel: -1,
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

func (a *ArcCachedDepNode) AsString() string {
	return fmt.Sprintf("%v h:%d l:%d left/right (mod,lset): (%v %v)/(%v %v)", a.String(), a.Head, a.ELabel, a.leftMods, a.leftLabels, a.rightMods, a.rightLabels)
}

func (a *ArcCachedDepNode) Equal(otherEq util.Equaler) bool {
	other := otherEq.(*ArcCachedDepNode)
	return reflect.DeepEqual(a.Node, other.Node) &&
		reflect.DeepEqual(a.leftMods, other.leftMods) &&
		reflect.DeepEqual(a.rightMods, other.rightMods)
}

func (a *ArcCachedDepNode) CopyArraySlice(aSrc, aDst *[3]int, sSrc, sDst *[]int) {
	if len(*sSrc) > cap(*aSrc) {
		*sDst = make([]int, len(*sSrc))
		copy(*sDst, *sSrc)
	} else {
		*sDst = (*aDst)[:len(*sSrc)]
	}
}

func (a *ArcCachedDepNode) Copy() *ArcCachedDepNode {
	aDst := new(ArcCachedDepNode)
	*aDst = *a
	a.CopyArraySlice(&a.leftModArray, &aDst.leftModArray, &a.leftMods, &aDst.leftMods)
	a.CopyArraySlice(&a.rightModArray, &aDst.rightModArray, &a.rightMods, &aDst.rightMods)
	a.CopyArraySlice(&a.leftLabelArray, &aDst.leftLabelArray, &a.leftLabels, &aDst.leftLabels)
	a.CopyArraySlice(&a.rightLabelArray, &aDst.rightLabelArray, &a.rightLabels, &aDst.rightLabels)
	return aDst
}
