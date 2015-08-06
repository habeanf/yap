package ma

import (
	"yap/alg/graph"
	"yap/nlp/format/lex"
	. "yap/nlp/types"
	"yap/util"

	"fmt"
	"log"
)

const ESTIMATED_MORPHS_PER_TOKEN = 5

type BGULex struct {
	MaxPrefixLen int
	Prefixes     map[string][]BasicMorphemes

	Lex map[string][]BasicMorphemes

	Files []string
	Stats *AnalyzeStats
}

var _ MorphologicalAnalyzer = &BGULex{}

func (l *BGULex) loadTokens(file, format string) {
	tokens, err := lex.ReadFile(file, format)
	if err != nil {
		panic(fmt.Sprintf("Failed to load %v: %v", file, err))
	}
	var m map[string][]BasicMorphemes
	if format == "prefix" {
		l.Prefixes = make(map[string][]BasicMorphemes, len(tokens))
		m = l.Prefixes
	} else if format == "lexicon" {
		l.Lex = make(map[string][]BasicMorphemes, len(tokens))
		m = l.Lex
	}
	log.Println("Found", len(tokens), "tokens in lexicon file:", file)
	for _, token := range tokens {
		if cur, exists := m[token.Token]; exists {
			m[token.Token] = append(cur, token.Morphemes...)
		} else {
			m[token.Token] = token.Morphemes
		}
	}
}

func (l *BGULex) LoadPrefixes(file string) {
	l.loadTokens(file, "prefix")
	l.MaxPrefixLen = 0
	for _, morphs := range l.Prefixes {
		if l.MaxPrefixLen < len(morphs) {
			l.MaxPrefixLen = len(morphs)
		}
	}
	log.Println("Loaded", len(l.Prefixes), "prefixes from lexicon")
}

func (l *BGULex) LoadLex(file string) {
	l.loadTokens(file, "lexicon")
	log.Println("Loaded", len(l.Lex), "tokens from lexicon")
}

func (l *BGULex) OOVAnalysis(input string) []BasicMorphemes {
	return []BasicMorphemes{BasicMorphemes([]*Morpheme{
		&Morpheme{
			BasicDirectedEdge: graph.BasicDirectedEdge{0, 0, 1},
			Form:              input,
			Lemma:             "",
			CPOS:              "NNP",
			POS:               "NNP",
			FeatureStr:        "",
		},
	})}
}
func (l *BGULex) AnalyzeToken(input string, startingNode, numToken int) (*Lattice, interface{}) {
	// log.Println("Analyzing token", numToken, "starting at", startingNode)
	lat := &Lattice{
		Token:     Token(input),
		Morphemes: make(Morphemes, 0, ESTIMATED_MORPHS_PER_TOKEN),
		Next:      make(map[int][]int, ESTIMATED_MORPHS_PER_TOKEN),
		BottomId:  startingNode,
		TopId:     startingNode,
	}
	lat.Next[0] = make([]int, 0, 1)
	var (
		prefixLat, hostLat                  []BasicMorphemes
		prefixExists, hostExists, anyExists bool
	)
	hostLat, hostExists = l.Lex[input]
	if hostExists {
		// log.Println("\tPrefix 0")
		lat.AddAnalysis(nil, hostLat, numToken)
		anyExists = true
	}
	for i := 1; i < util.Min(l.MaxPrefixLen, len(input)); i++ {
		prefixLat, prefixExists = l.Prefixes[input[:i]]
		if prefixExists {
			hostLat, hostExists = l.Lex[input[i:]]
			if hostExists {
				// log.Println("\tPrefix", i)
				lat.AddAnalysis(prefixLat, hostLat, numToken)
				anyExists = true
			}
		}
	}
	if !anyExists {
		log.Println("Token is OOV:", input)
		if l.Stats != nil {
			l.Stats.OOVTokens++
			l.Stats.AddOOVToken(input)
		}
		hostLat = l.OOVAnalysis(input)
		lat.AddAnalysis(nil, hostLat, numToken)
	}
	lat.Optimize()
	return lat, nil
}

func (l *BGULex) Analyze(input []string) (LatticeSentence, interface{}) {
	retval := make(LatticeSentence, len(input))
	var (
		lat     *Lattice
		curNode int
	)
	for i, token := range input {
		if l.Stats != nil {
			l.Stats.TotalTokens++
			l.Stats.AddToken(token)
		}
		lat, _ = l.AnalyzeToken(token, curNode, i)
		curNode = lat.Top()
		// log.Println("New top is", curNode)
		retval[i] = *lat
	}
	return retval, nil
}
