package Types

import "chukuparser/Algorithm/Graph"

type Morpheme struct {
	Form     string
	POS      string
	Features map[string]string
}

type Spellout []Morpheme

type Lattice struct {
	Token     Token
	Morphemes []Morpheme
	Edges     []Graph.BasicDirectedEdge
}
