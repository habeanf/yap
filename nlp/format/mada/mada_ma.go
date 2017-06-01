package mada

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"yap/alg/graph"
	nlp "yap/nlp/types"
)

// Package mada reads morphological analyses output by MADA

const (
	PREF_SENT  = ";;; SENTENCE"
	PREF_WORD  = ";;WORD"
	PREF_PRED  = ";;SVM_PREDICTIONS:"
	SENT_BREAK = "SENTENCE BREAK"
	LINE_SEP   = "--------------"

	FIELD_SEP   = " "
	KV_SEP      = ":"
	BW_WORD_SEP = "+"
	BW_TAG_SEP  = "/"
)

type Analysis struct {
	Pref     byte // *|^|_
	Score    float32
	FieldMap map[string]string
	Fields   []string
	BW       [][2]string // buckwalter rep
}

type Word struct {
	Token    string
	Pred     []string
	Analyses []Analysis
}

type Sentence struct {
	Tokens []string
	Words  []Word
}

func Read(reader io.Reader, limit int) ([]*Sentence, error) {
	var (
		sentences   []*Sentence
		currentSent *Sentence
	)
	bufReader := bufio.NewReader(reader)

	for curLine, isPrefix, err := bufReader.ReadLine(); err == nil; curLine, isPrefix, err = bufReader.ReadLine() {
		if isPrefix {
			panic("Buffer not large enough, fix me :(")
		}
		buf := bytes.NewBuffer(curLine).String()
		if buf == LINE_SEP {
			continue
		}
		if buf == SENT_BREAK {
			sentences = append(sentences, currentSent)
			if len(sentences) >= limit {
				break
			}
			continue
		}
		record := strings.Split(buf, FIELD_SEP)
		if strings.HasPrefix(buf, PREF_SENT) {
			currentSent = &Sentence{
				Tokens: record[2:],
				Words:  make([]Word, 0, len(record)-2),
			}
			continue
		}
		if strings.HasPrefix(buf, PREF_WORD) {
			currentSent.Words = append(currentSent.Words, Word{
				Token:    record[1],
				Pred:     nil,
				Analyses: nil,
			})
			continue
		}
		curWord := &currentSent.Words[len(currentSent.Words)-1]
		if strings.HasPrefix(buf, PREF_PRED) {
			curWord.Pred = record[1:]
			continue
		}

		rowScore, err := strconv.ParseFloat(record[0][1:], 32)
		if err != nil {
			panic("Got unparsable score")
		}
		// row must be a spellout
		analysis := Analysis{
			Pref:     byte(record[0][0]),
			Score:    float32(rowScore),
			Fields:   record[1:],
			FieldMap: make(map[string]string, len(record)-1),
		}
		// find buckwalter string while setting k:v fields of analyses
		var bwStr string
		for _, field := range record[1:] {
			if strings.HasPrefix(field, "bw:") {
				bwStr = field[3:]
			}
			kvStrs := strings.Split(field, KV_SEP)
			analysis.FieldMap[kvStrs[0]] = kvStrs[1]
		}
		for _, morph := range strings.Split(bwStr, BW_WORD_SEP) {
			if len(morph) == 0 {
				continue
			}
			morphSplit := strings.Split(morph, BW_TAG_SEP)
			analysis.BW = append(analysis.BW, [2]string{morphSplit[0], morphSplit[1]})
		}
		curWord.Analyses = append(curWord.Analyses, analysis)
	}
	return sentences, nil
}

func ReadFile(filename string, limit int) ([]*Sentence, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	return Read(file, limit)
}

func Word2Lattice(word Word, startNode, tokenNum int) *nlp.Lattice {
	wordLattice := &nlp.Lattice{
		Token:    nlp.Token(word.Token),
		Next:     make(map[int][]int),
		BottomId: startNode,
		TopId:    startNode,
	}
	for i, analysis := range word.Analyses {
		log.Println("At analysis", i)
		basicMorphs := make(nlp.BasicMorphemes, len(analysis.BW))
		for j, madaMorph := range analysis.BW {
			morph := &nlp.Morpheme{
				BasicDirectedEdge: graph.BasicDirectedEdge{0, j, j + 1},
				Form:              madaMorph[0],
				CPOS:              madaMorph[1],
			}
			basicMorphs[j] = morph
		}
		wordLattice.AddAnalysis(nil, []nlp.BasicMorphemes{basicMorphs}, tokenNum)
	}

	return wordLattice
}

func MADA2Lattice(sent *Sentence) nlp.LatticeSentence {
	lat := make(nlp.LatticeSentence, len(sent.Words))
	var startNode int
	for i, word := range sent.Words {
		log.Println("At word", i)
		wordLattice := Word2Lattice(word, startNode, i)
		startNode = wordLattice.Top()
		lat[i] = *wordLattice
	}
	return lat
}

func MADA2LatticeCorpus(sents []*Sentence) []nlp.LatticeSentence {
	corpus := make([]nlp.LatticeSentence, len(sents))
	for i, sent := range sents {
		corpus[i] = MADA2Lattice(sent)
	}
	return corpus
}
