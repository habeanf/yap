package Transition

import (
	"chukuparser/Algorithm/Graph"
	"chukuparser/Algorithm/Transition"
	"chukuparser/NLP"
	"chukuparser/Util"
	"fmt"
	"reflect"
	"strings"
)

type DependencyConfiguration interface {
	Util.Equaler
	Conf() Transition.Configuration
	Graph() NLP.LabeledDependencyGraph
	Address(location []byte) (int, bool)
	Attribute(nodeID int, attribute []byte) (string, bool)
	Previous() DependencyConfiguration
}

type TaggedDepNode struct {
	Id    int
	Token string
	POS   string
}

var _ NLP.DepNode = &TaggedDepNode{}

func (t *TaggedDepNode) ID() int {
	return t.Id
}

func (t *TaggedDepNode) String() string {
	return t.Token
}

func (t *TaggedDepNode) Equal(otherEq Util.Equaler) bool {
	other := otherEq.(*TaggedDepNode)
	return reflect.DeepEqual(t, other)
}

type BasicDepArc struct {
	Head     int
	Relation NLP.DepRel
	Modifier int
}

var _ NLP.LabeledDepArc = &BasicDepArc{}

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
	return arc.Relation
}

func (arc *BasicDepArc) Equal(otherEq Util.Equaler) bool {
	other := otherEq.(*BasicDepArc)
	return reflect.DeepEqual(arc, other)
}

func (arc *BasicDepArc) String() string {
	return fmt.Sprintf("(%d,%s,%d)", arc.GetHead(), arc.GetRelation(), arc.GetModifier())
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
		if g.NumberOfNodes() != other.NumberOfNodes() {
			fmt.Println("\tNumber of nodes are not equal")
		}
		if g.NumberOfArcs() != other.NumberOfArcs() {
			// fmt.Println("\tNumber of arcs are not equal")
		}
		// fmt.Println("Nodes (Gold,Actual)", g.NumberOfNodes(), other.NumberOfNodes())
		// fmt.Println("Arcs (Gold,Actual)", g.NumberOfArcs(), other.NumberOfArcs())
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
	// if !nodesEqual {
	// 	// fmt.Println("\tNodes not equal")
	// 	// fmt.Println(g.Nodes)
	// 	// fmt.Println(nodes)
	// }
	// if !arcsEqual {
	// 	// fmt.Print("\tArcs diff (left,right) ")
	// 	// diffLeft, diffRight := gArcSet.Diff(otherArcSet)
	// 	// fmt.Printf("(%v,%v)\n", diffLeft.Size(), diffRight.Size())
	// 	// sortLeft := gArcSet.Sorted()
	// 	// sortRight := otherArcSet.Sorted()
	// 	// for i := 0; i < Util.Max(sortLeft.Len(), sortRight.Len()); i++ {
	// 	// 	fmt.Print("\t")
	// 	// 	if i < sortLeft.Len() {
	// 	// 		fmt.Print(sortLeft.arcset[i])
	// 	// 	}
	// 	// 	fmt.Print("\t")
	// 	// 	if i < sortRight.Len() {
	// 	// 		fmt.Print(sortRight.arcset[i])
	// 	// 	}
	// 	// 	fmt.Print("\n")
	// 	// }
	// }
	// if numArcsNotEqual > 0 {
	// 	fmt.Println("\t", numArcsNotEqual, "Arcs not equal")
	// 	fmt.Println("\t", g.Arcs)
	// 	fmt.Println("\t", arcs)
	// }
	return nodesEqual && arcsEqual
}

func (g *BasicDepGraph) Sentence() NLP.Sentence {
	return NLP.Sentence(g.TaggedSentence())
}

func (g *BasicDepGraph) TaggedSentence() NLP.TaggedSentence {
	sent := make([]NLP.TaggedToken, g.NumberOfNodes()-1)
	for _, node := range g.Nodes {
		taggedNode := node.(*TaggedDepNode)
		if taggedNode.Token == ROOT_TOKEN {
			continue
		}
		sent[taggedNode.ID()-2] = NLP.TaggedToken{taggedNode.Token, taggedNode.POS}
	}
	return NLP.TaggedSentence(NLP.BasicTaggedSentence(sent))
}
