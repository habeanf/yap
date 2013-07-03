package Lattices

type LatticeRow struct {
	Start int
	End   int
	Word  string
	Lemma string
	CPos  string
	Pos   string
	Morph map[string][]string
	Token int
}

type Lattice []LatticeRow
