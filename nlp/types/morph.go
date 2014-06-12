package types

import (
	"chukuparser/algorithm"
	"chukuparser/algorithm/graph"
	"chukuparser/util"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
)

type Morpheme struct {
	graph.BasicDirectedEdge
	Form       string
	CPOS       string
	POS        string
	Features   map[string]string
	TokenID    int
	FeatureStr string
}

type EMorpheme struct {
	Morpheme
	EForm, EFCPOS, EPOS int
	EFeatures           int
}

var _ DepNode = &Morpheme{}
var _ DepNode = &EMorpheme{}

func NewRootMorpheme() *EMorpheme {
	return &EMorpheme{Morpheme: Morpheme{
		graph.BasicDirectedEdge{0, 0, 0},
		ROOT_TOKEN, ROOT_TOKEN, ROOT_TOKEN,
		nil, 0, "",
	}}
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

func (m *Morpheme) String() string {
	return fmt.Sprintf("%v-%v-%v-%s", m.Form, m.CPOS, m.POS, m.FeatureStr)
}

func (m *Morpheme) Equal(otherEq util.Equaler) bool {
	other := otherEq.(*Morpheme)
	return m.Form == other.Form &&
		m.CPOS == other.CPOS &&
		m.POS == other.POS &&
		reflect.DeepEqual(m.Features, other.Features)
}

func (m *EMorpheme) Equal(otherEq util.Equaler) bool {
	other := otherEq.(*EMorpheme)
	return m.Form == other.Form &&
		m.CPOS == other.CPOS &&
		m.POS == other.POS &&
		reflect.DeepEqual(m.Features, other.Features)
}

func (m *EMorpheme) Copy() *EMorpheme {
	newMorph := new(EMorpheme)
	*newMorph = *m
	newMorph.Features = make(map[string]string)
	for k, v := range m.Features {
		newMorph.Features[k] = v
	}
	return newMorph
}

var _ graph.DirectedEdge = &Morpheme{}
var _ graph.DirectedEdge = &EMorpheme{}

type Morphemes []*EMorpheme

func (m Morphemes) Len() int {
	return len(m)
}

func (m Morphemes) Less(i, j int) bool {
	return m[i].From() < m[j].From() ||
		(m[i].From() == m[j].From() && m[i].To() < m[j].To())
}

func (m Morphemes) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type Spellout Morphemes

type Spellouts []Spellout

func (s Spellouts) Len() int {
	return len(s)
}

func (s Spellouts) Less(i, j int) bool {
	return s[i].AsString() < s[j].AsString()
}

func (s Spellouts) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Spellout) String() string {
	posStrings := make([]string, len(s))
	for i, morph := range s {
		posStrings[i] = morph.CPOS
	}
	return strings.Join(posStrings, ":")
}

func (s Spellout) AsString() string {
	strs := make([]string, len(s))
	for i, morph := range s {
		strs[i] = morph.String()
	}
	return strings.Join(strs, ";")
}

func (s Spellout) Equal(other Spellout) bool {
	if len(s) != len(other) {
		return false
	}
	for i, val := range other {
		if !s[i].Equal(val) {
			return false
		}
	}
	return true
}

func (s Spellouts) Find(other Spellout) (int, bool) {
	for i, cur := range s {
		if cur.Equal(other) {
			return i, true
		}
	}
	return 0, false
}

type Mapping struct {
	Token    Token
	Spellout Spellout
}

func (m *Mapping) Equal(other *Mapping) bool {
	return m.Token == other.Token && m.Spellout.Equal(other.Spellout)
}

type Mappings []*Mapping

var _ algorithm.Index = make(Mappings, 1)

func (ms Mappings) Equal(otherEq util.Equaler) bool {
	other, ok := otherEq.(Mappings)
	if !ok {
		return false
	}
	if len(ms) != len(other) {
		return false
	}
	for i, m := range ms {
		if !m.Equal(other[i]) {
			return false
		}
	}
	return true
}

func (ms Mappings) Index(i int) (int, bool) {
	if i >= len(ms) {
		return 0, false
	}
	return i, true
}

type Path int

type Lattice struct {
	Token           Token
	Morphemes       Morphemes
	Spellouts       Spellouts
	Next            map[int][]int
	BottomId, TopId int
}

func (l *Lattice) BridgeMissingMorphemes() {
	for _, m := range l.Morphemes {
		if _, exists := l.Next[m.To()]; !exists && m.To() < l.TopId {
			if _, nextExists := l.Next[m.To()+1]; nextExists {
				log.Println("Bridging morpheme", m.Form, "from", m.To(), "to", m.To()+1)
				m.BasicDirectedEdge[2] += 1
			} else {
				log.Println("Morpheme's next does not exist and cannot bridge! (", m.Form, m.From(), m.To(), ")")
			}
		}
	}
}

func (l *Lattice) UnionPath(other *Lattice) {
	// assume other is a "gold" path (only one "next" at each node)
	// add gold lattice path if it is an alternative to existing paths with the
	// same nodes
	formMorphs := make(map[string][]*EMorpheme)
	for _, predMorph := range l.Morphemes {
		if cur, exists := formMorphs[predMorph.Form]; exists {
			formMorphs[predMorph.Form] = append(cur, predMorph)
		} else {
			formMorphs[predMorph.Form] = []*EMorpheme{predMorph}
		}
	}
	var found bool
	for _, goldMorph := range other.Morphemes {
		if curMorphs, exists := formMorphs[goldMorph.Form]; exists {
			for _, curMorph := range curMorphs {
				if curMorph.Equal(goldMorph) {
					found = true
				}
			}
		} else {
			log.Println("Warning: gold morph form", goldMorph.Form, "is not in pred lattice!")
			continue
		}
		if !found {
			newMorph := goldMorph.Copy()
			log.Println("Adding missing morpheme", goldMorph.Form, goldMorph.POS, goldMorph.CPOS, goldMorph.FeatureStr)
			exampleMorphs, _ := formMorphs[goldMorph.Form]
			exampleMorph := exampleMorphs[0]
			newMorph.Morpheme.BasicDirectedEdge[1] = exampleMorph.From()
			newMorph.Morpheme.BasicDirectedEdge[2] = exampleMorph.To()
			id := len(l.Morphemes)
			l.Morphemes = append(l.Morphemes, newMorph)
			mList, _ := l.Next[newMorph.From()]
			l.Next[newMorph.From()] = append(mList, id)
		}
		found = false
	}
}

func NewRootLattice() Lattice {
	morphs := make(Morphemes, 1)
	morphs[0] = NewRootMorpheme()
	lat := &Lattice{
		ROOT_TOKEN,
		morphs,
		nil,
		make(map[int][]int),
		0,
		0,
	}
	return *lat
}

type LatticeSentence []Lattice

var _ Sentence = LatticeSentence{}

func (ls LatticeSentence) Tokens() []string {
	res := make([]string, len(ls))
	for i, val := range ls {
		res[i] = string(val.Token)
	}
	return res
}

func (ls LatticeSentence) Equal(otherEq util.Equaler) bool {
	otherSent := otherEq.(Sentence)
	if len(otherSent.Tokens()) != len(ls) {
		return false
	}
	otherToks := otherSent.Tokens()
	curToks := ls.Tokens()
	return reflect.DeepEqual(curToks, otherToks)
}

func (l *Lattice) GetDirectedEdge(i int) graph.DirectedEdge {
	return graph.DirectedEdge(l.Morphemes[i])
}

func (l *Lattice) GetEdge(i int) graph.Edge {
	return graph.Edge(l.Morphemes[i])
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

func (l *Lattice) GetVertex(i int) graph.Vertex {
	return graph.BasicVertex(i)
}

func (l *Lattice) NumberOfEdges() int {
	return len(l.Morphemes)
}

func (l *Lattice) NumberOfVertices() int {
	return l.Top() - l.Bottom()
}

var _ graph.BoundedLattice = &Lattice{}
var _ graph.DirectedGraph = &Lattice{}

// untested..
func (l *Lattice) Inf(i, j int) int {
	iReachable := make(map[int]int)
	for path := range graph.YieldAllPaths(graph.DirectedGraph(l), l.Bottom(), i) {
		for i, el := range path {
			dist := len(path) - i - 1
			iReachable[el.ID()] = dist
		}
	}
	var bestVal, bestDist int = -1, -1
	for path := range graph.YieldAllPaths(graph.DirectedGraph(l), l.Bottom(), j) {
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
	for path := range graph.YieldAllPaths(graph.DirectedGraph(l), i, l.Top()) {
		for dist, el := range path {
			iReachable[el.ID()] = dist
		}
	}
	var bestVal, bestDist int = -1, -1
	for path := range graph.YieldAllPaths(graph.DirectedGraph(l), j, l.Top()) {
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
	return l.TopId
}

func (l *Lattice) Bottom() int {
	return l.BottomId
}

func (l *Lattice) MaxPathLen() int {
	if len(l.Morphemes) == 0 {
		return 0
	}
	return l.Top() - l.Bottom()
}

func (l *Lattice) SortMorphemes() {
	sort.Sort(l.Morphemes)
}

func (l *Lattice) SortNexts() {
	for _, next := range l.Next {
		sort.Ints(next)
	}
}

func (l *Lattice) GenToken() {
	if l.Spellouts == nil {
		panic("Can't generate token without a spellout")
	}
	if len(l.Spellouts) == 0 {
		l.Token = Token("")
		return
	}
	spellout := l.Spellouts[0]
	strs := make([]string, len(spellout))
	for i, morph := range spellout {
		strs[i] = morph.Form
	}
	l.Token = Token(strings.Join(strs, ""))
}

func (l *Lattice) GenSpellouts() {
	if l.Spellouts != nil {
		return
	}
	if len(l.Morphemes) == 0 {
		l.Spellouts = make(Spellouts, 0)
		return
	}
	var (
		pathId   int
		from, to int = l.Bottom(), l.Top()
	)
	l.Spellouts = make(Spellouts, 0, to-from)
	for path := range graph.YieldAllPaths(graph.DirectedGraph(l), from, to) {
		spellout := make(Spellout, len(path))
		for i, el := range path {
			spellout[i] = el.(*EMorpheme)
		}
		l.Spellouts = append(l.Spellouts, spellout)

		pathId++
	}
	sort.Sort(l.Spellouts)
}

func (l *Lattice) YieldPaths() chan Path {
	l.GenSpellouts()
	pathChan := make(chan Path)
	go func() {
		for i, _ := range l.Spellouts {
			pathChan <- Path(i)
		}
		close(pathChan)
	}()
	return pathChan
}

func (l *Lattice) Path(i int) Spellout {
	return l.Spellouts[i]
}

type MorphDependencyGraph interface {
	LabeledDependencyGraph
	GetMappings() []*Mapping
	GetMorpheme(int) *EMorpheme
}
