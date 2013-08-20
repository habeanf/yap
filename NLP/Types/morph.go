package Types

import (
	"chukuparser/Algorithm/Graph"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Morpheme struct {
	Graph.BasicDirectedEdge
	Form     string
	CPOS     string
	POS      string
	Features map[string]string
	TokenID  int
}

func (m *Morpheme) ID() int {
	return m.BasicDirectedEdge.ID()
}

func (m *Morpheme) From() int {
	return m.BasicDirectedEdge.From()
}

func (m *Morpheme) To() int {
	return m.BasicDirectedEdge.To()
}

var _ Graph.DirectedEdge = &Morpheme{}

type Spellout []*Morpheme

type Spellouts []Spellout

func (s Spellout) AsString() string {
	posStrings := make([]string, len(s))
	for i, morph := range s {
		posStrings[i] = morph.CPOS
	}
	return strings.Join(posStrings, "-")
}

func (s Spellout) String() string {
	strs := make([]string, len(s))
	for i, morph := range s {
		strs[i] = fmt.Sprintf("%v", morph)
	}
	return fmt.Sprintf("%v", strs)
}

type Path string

type Lattice struct {
	Token     Token
	Morphemes []*Morpheme
	Spellouts Spellouts
}

func (s Spellouts) Find(other Spellout) (int, bool) {
	for i, cur := range s {
		if len(cur) == len(s) && reflect.DeepEqual(cur, other) {
			return i, true
		}
	}
	return 0, false
}

func (l *Lattice) GetDirectedEdge(i int) Graph.DirectedEdge {
	return Graph.DirectedEdge(l.Morphemes[i])
}

func (l *Lattice) GetEdge(i int) Graph.Edge {
	return Graph.Edge(l.Morphemes[i])
}

func (l *Lattice) GetEdges() []int {
	res := make([]int, len(l.Morphemes))
	for i, _ := range l.Morphemes {
		res[i] = i
	}
	return res
}

func (l *Lattice) GetVertices() []int {
	vSet := make(map[int]bool)
	for _, edge := range l.Morphemes {
		vSet[edge.From()] = true
		vSet[edge.To()] = true
	}
	res := make([]int, 0, len(vSet))
	for k, _ := range vSet {
		res = append(res, k)
	}
	return res
}

func (l *Lattice) GetVertex(i int) Graph.Vertex {
	return Graph.BasicVertex(i)
}

func (l *Lattice) NumberOfEdges() int {
	return len(l.Morphemes)
}

func (l *Lattice) NumberOfVertices() int {
	return l.Top() - l.Bottom()
}

var _ Graph.BoundedLattice = &Lattice{}
var _ Graph.DirectedGraph = &Lattice{}

// untested..
func (l *Lattice) Inf(i, j int) int {
	iReachable := make(map[int]int)
	for path := range Graph.YieldAllPaths(Graph.DirectedGraph(l), l.Bottom(), i) {
		for i, el := range path {
			dist := len(path) - i - 1
			iReachable[el.ID()] = dist
		}
	}
	var bestVal, bestDist int = -1, -1
	for path := range Graph.YieldAllPaths(Graph.DirectedGraph(l), l.Bottom(), j) {
		for i, _ := range path {
			el := path[len(path)-i-1]
			dist, exists := iReachable[el.ID()]
			if exists {
				if bestDist == -1 || bestDist > dist {
					bestVal = el.ID()
					bestDist = dist
					break
				}
			}
		}
	}
	return bestVal
}

// untested..
func (l *Lattice) Sup(i, j int) int {
	iReachable := make(map[int]int)
	for path := range Graph.YieldAllPaths(Graph.DirectedGraph(l), i, l.Top()) {
		for dist, el := range path {
			iReachable[el.ID()] = dist
		}
	}
	var bestVal, bestDist int = -1, -1
	for path := range Graph.YieldAllPaths(Graph.DirectedGraph(l), j, l.Top()) {
		for _, el := range path {
			dist, exists := iReachable[el.ID()]
			if exists {
				if bestDist == -1 || bestDist > dist {
					bestVal = el.ID()
					bestDist = dist
					break
				}
			}
		}
	}
	return bestVal
}

func (l *Lattice) Top() int {
	return l.Morphemes[len(l.Morphemes)-1].To()
}

func (l *Lattice) Bottom() int {
	return l.Morphemes[0].From()
}

func (l *Lattice) GenSpellouts() {
	if l.Spellouts != nil {
		return
	}
	var (
		pathId   int
		from, to int = l.Bottom(), l.Top()
	)
	l.Spellouts = make(Spellouts, 0, to-from)
	for path := range Graph.YieldAllPaths(Graph.DirectedGraph(l), from, to) {
		spellout := make(Spellout, len(path))
		for i, el := range path {
			spellout[i] = el.(*Morpheme)
		}
		l.Spellouts = append(l.Spellouts, spellout)

		pathId++
	}
}

func (l *Lattice) YieldPaths() chan Path {
	l.GenSpellouts()
	pathChan := make(chan Path)
	go func() {
		for i, _ := range l.Spellouts {
			pathChan <- Path(strconv.Itoa(i))
		}
	}()
	return pathChan
}

func (l *Lattice) Path(i int) Spellout {
	return l.Spellouts[i]
}
