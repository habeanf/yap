package Transition

import (
	"chukuparser/NLP"
	"testing"
)

func TestTaggedDepNode(t *testing.T) {
	node := &TaggedDepNode{0, "token", "tag"}
	if node.ID() != 0 {
		t.Error("Got wrong ID")
	}
	if node.Token != "token" {
		t.Error("Wrong token value")
	}
	if node.POS != "tag" {
		t.Error("Wrong tag value")
	}
	if node.String() != "token" {
		t.Error("Got wrong String representation")
	}
	other := node
	if !node.Equal(other) {
		t.Error("Failed equality on equal pointers")
	}
	other = &TaggedDepNode{0, "token", "tag2"}
	if node.Equal(other) {
		t.Error("Returned equal on non-equal nodes")
	}
	other.POS = "tag"
	if !node.Equal(other) {
		t.Error("Returned not equal on equal by value")
	}
}

func TestBasicDepArc(t *testing.T) {
	arc := &BasicDepArc{1, NLP.DepRel("rel"), 5}
	vertices := arc.Vertices()
	if len(vertices) != 2 {
		t.Error("Wrong number of Vertices")
	}
	if vertices[0] != 1 {
		t.Error("Wrong head in Vertices")
	}
	if vertices[1] != 5 {
		t.Error("Wrong modifier in Vertices")
	}
	if arc.From() != 5 {
		t.Error("Wrong from vertex")
	}
	if arc.To() != 1 {
		t.Error("Wrong to vertex")
	}
	if arc.GetHead() != 1 {
		t.Error("Wrong head")
	}
	if arc.GetModifier() != 5 {
		t.Error("Wrong modifier")
	}
	if arc.GetRelation() != NLP.DepRel("rel") {
		t.Error("Wrong relation")
	}
}

func TestBasicDepGraph(t *testing.T) {
	g := &BasicDepGraph{[]NLP.DepNode{}, []*BasicDepArc{}}
	if g.NumberOfNodes() != 0 ||
		g.NumberOfArcs() != 0 ||
		g.NumberOfEdges() != 0 ||
		g.NumberOfVertices() != 0 {
		t.Error("Got wrong number of Nodes/Arcs/Edges/Vertices for empty graph")
	}
	if len(g.GetEdges()) != 0 {
		t.Error("Got non empty edge index slice for empty graph")
	}
	if g.GetVertex(0) != nil ||
		g.GetEdge(0) != nil ||
		g.GetNode(0) != nil ||
		g.GetArc(0) != nil {
		t.Error("Got non-nil edge/vertex/arc/node for empty graph")
	}
	g = &BasicDepGraph{
		[]NLP.DepNode{&TaggedDepNode{0, "v1", "tag1"},
			&TaggedDepNode{1, "v1", "tag2"}},
		[]*BasicDepArc{&BasicDepArc{0, "a", 1}}}
	if g.NumberOfNodes() != 2 || g.NumberOfVertices() != 2 {
		t.Error("Got wrong number of nodes/vertices")
	}
	if g.NumberOfEdges() != 1 || g.NumberOfArcs() != 1 {
		t.Error("Got wrong number of arcs/edges")
	}
	if len(g.GetVertices()) != 2 {
		t.Error("Got wrong number of vertex indices")
	}
	if len(g.GetEdges()) != 1 {
		t.Error("Got wrong number of edge indices")
	}
	if g.GetVertex(0) != g.Nodes[0] {
		t.Error("Got wrong vertex")
	}
	if g.GetVertex(1) != g.Nodes[1] {
		t.Error("Got wrong vertex")
	}
	if g.GetNode(0) != g.Nodes[0] {
		t.Error("Got wrong node")
	}
	if g.GetNode(1) != g.Nodes[1] {
		t.Error("Got wrong node")
	}
	if g.GetArc(0) != g.Arcs[0] {
		t.Error("Got wrong arc")
	}
	if g.GetEdge(0) != g.Arcs[0] {
		t.Error("Got wrong edge")
	}
	if g.GetDirectedEdge(0) != g.Arcs[0] {
		t.Error("Got wrong directed edge")
	}
	if g.GetLabeledArc(0) != g.Arcs[0] {
		t.Error("Got wrong labeled arc")
	}
	if len(g.StringEdges()) == 0 {
		t.Error("Got empty StringEdges()")
	}
}
