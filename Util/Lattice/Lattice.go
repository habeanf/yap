package Lattice

type Row struct {
	Start int
	End   int
	Word  string
	Lemma string
	CPos  string
	Pos   string
	Morph map[string][]string
	Token int
}

type Lattice map[int]Row

type Lattices []Lattice
