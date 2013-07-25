package Transition

type TaggedDepNode struct {
	ID    string
	Token string
	POS   string
}

var _ DepNode = TaggedDepNode

func (t *TaggedDepNode) ID() int {
	return t.ID
}

func (t *TaggedDepNode) String() string {
	return t.Token
}

func (t *TaggedDepNode) GetProperty(prop string) (string, bool) {
	switch prop {
	case "w":
		return t.Token(), true
	case "p":
		return t.POS, true
	default:
		return "", false
	}
}

type BasicDepArc struct {
	Modifier int
	Relation DepRel
	Head     int
}

var _ DepArc = BasicDepArc{}

func (arc *BasicDepArc) Vertices() []int {
	return []int{arc.Head, arc.Modifier}
}

func (arc *BasicDepArc) From() int {
	return arc.Modifier
}

func (arc *BasicDepArc) To() int {
	return arc.Head
}

func (arc *BasicDepArc) GetProperty(property string) (string, bool) {
	if property == "l" {
		return arc.Relation, true
	} else {
		return "", false
	}
}

type BasicDepGraph struct {
	Nodes []*DepNode
	Arcs  []*BasicDepArc
}

// Verify BasicDepGraph is a labeled dep. graph
var _ DependencyGraph = BasicDepGraph{}
var _ Labeled = BasicDepGraph

func (g *BasicDepGraph) GetVertices() []int {
	retval := make([]int, len(g.Nodes))
	for i := 0; i < len(g.Nodes); i++ {
		retval[i] = i
	}
	return retval
}

func (g *BasicDepGraph) GetEdges() []int {
	retval := make([]int, len(g.Edges))
	for i := 0; i < len(g.Edges); i++ {
		retval[i] = i
	}
	return retval
}

func (g *BasicDepGraph) GetVertices() []int {
	retval := make([]int, len(g.Nodes))
	for i := 0; i < len(g.Nodes); i++ {
		retval[i] = i
	}
	return retval
}

func (g *BasicDepGraph) GetVertex(n int) *Vertex {
	return g.Nodes[n]
}

func (g *BasicDepGraph) GetEdge(n int) *Edge {
	return g.Arcs[n]
}

func (g *BasicDepGraph) GetEdge(n int) *DirectedEdge {
	return g.Arcs[n]
}

func (g *BasicDepGraph) NumberOfVertices() int {
	return len(g.Nodes)
}

func (g *BasicDepGraph) NumberOfEdges() int {
	return len(g.Arcs)
}

func (g *BasicDepGraph) GetNode(n int) *DepNode {
	return g.Nodes[n]
}

func (g *BasicDepGraph) GetArc(n int) *DepArc {
	return g.Arcs[n]
}
