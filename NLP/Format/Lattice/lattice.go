package Lattice

// Package Lattice reads lattice format files

import (
	"chukuparser/NLP"

	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Features map[string]string

func (f Features) String() string {
	if f != nil || len(f) == 0 {
		return "_"
	}
	return fmt.Sprintf("%v", map[string]string(f))
}

type Edge struct {
	Start   int
	End     int
	Word    string
	Lemma   string
	CPosTag string
	PosTag  string
	Feats   Features
	Token   int
}

func (e Edge) String() string {
	fields := []string{
		fmt.Sprintf("%d", e.Start),
		fmt.Sprintf("%d", e.End),
		e.Word,
		"_",
		e.CPosTag,
		e.PosTag,
		e.Feats.String(),
		fmt.Sprintf("%d", e.Token),
		e.DepRel,
	}
	return strings.Join(fields, "\t")
}

type Lattice map[int][]Edge

type Lattices []Lattice

const (
	FIELD_SEPARATOR      = '\t'
	NUM_FIELDS           = 10
	FEATURES_SEPARATOR   = "|"
	FEATURE_SEPARATOR    = "="
	FEATURE_CONCAT_DELIM = ","
)

func ParseInt(value string) (int, error) {
	if value == "_" {
		return 0, nil
	}
	i, err := strconv.ParseInt(value, 10, 0)
	return int(i), err
}

func ParseString(value string) string {
	if value == "_" {
		return ""
	} else {
		return value
	}
}

func ParseFeatures(featuresStr string) (Features, error) {
	var featureMap Features
	if featuresStr == "_" {
		return featureMap, nil
	}

	featureList := strings.Split(featuresStr, FEATURES_SEPARATOR)
	if len(featureList) == 0 {
		return nil, errors.New("No features found, field should be '_'")
	}
	featureMap = make(Features, len(featureList))
	for _, featureStr := range featureList {
		featureKV := strings.Split(featureStr, FEATURE_SEPARATOR)
		if len(featureKV) != 2 {
			return nil, errors.New("Wrong number of fields for split of feature" + featureStr)
		}
		featName := featureKV[0]
		featValue := featureKV[1]
		existingFeatValue, featExist := featureMap[featName]
		if featExist {
			featureMap[featName] = existingFeatValue + FEATURE_CONCAT_DELIM + featValue
		} else {
			featureMap[featName] = featValue
		}
	}
	return featureMap, nil
}

func ParseEdge(record []string) (Edge, error) {
	var row Edge
	start, err := ParseInt(record[0])
	if err != nil {
		return row, errors.New(fmt.Sprintf("Error parsing START field (%s): %s", record[0], err.Error()))
	}
	row.Start = start

	end, err := ParseInt(record[1])
	if err != nil {
		return row, errors.New(fmt.Sprintf("Error parsing END field (%s): %s", record[1], err.Error()))
	}
	row.End = end

	word := ParseString(record[2])
	if word == "" {
		return row, errors.New("Empty WORD field")
	}
	row.Word = word

	cpostag := ParseString(record[4])
	if cpostag == "" {
		return row, errors.New("Empty CPOSTAG field")
	}
	row.CPosTag = cpostag

	postag := ParseString(record[5])
	if postag == "" {
		return row, errors.New("Empty POSTAG field")
	}
	row.PosTag = postag

	token, err := ParseInt(record[7])
	if err != nil {
		return row, errors.New(fmt.Sprintf("Error parsing TOKEN field (%s): %s", record[7], err.Error()))
	}
	row.Token = token

	features, err := ParseFeatures(record[6])
	if err != nil {
		return row, errors.New(fmt.Sprintf("Error parsing FEATS field (%s): %s", record[6], err.Error()))
	}
	row.Feats = features
	return row, nil
}

func Read(r io.Reader) ([]Lattice, error) {
	var sentences []Lattice
	reader := csv.NewReader(r)
	reader.Comma = FIELD_SEPARATOR
	reader.FieldsPerRecord = NUM_FIELDS

	records, err := reader.ReadAll()
	if err != nil {
		return nil, errors.New("Failure reading delimited file")
	}

	var currentLatt Lattice = nil
	for i, record := range records {
		// a record with id '1' indicates a new sentence
		// since csv reader ignores empty lines
		if record[0] == "1" {
			// store current sentence
			if currentLatt != nil {
				sentences = append(sentences, currentLatt)
			}
			currentLatt = make(Lattice)
		}

		edge, err := ParseEdge(record)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Error processing record %d at statement %d: %s", i, len(sentences), err.Error()))
		}
		edges, exists := currentLatt[edge.Start]
		if exists {
			currentLatt[edge.Start] = append(edges, edge)
		} else {
			currentLatt[edge.Start] = []Edge{edge}
		}
	}
	sentences = append(sentences, currentLatt)
	return sentences, nil
}

func Write(w io.Writer, lattices []Lattice) error {
	for _, lattice := range lattices {
		for i := 1; i < len(lattice); i++ {
			row := sent[i]
			for _, edge := range row {
				writer.Write(append([]byte(edge.String()), '\n'))
			}
		}
		writer.Write([]byte{'\n'})
	}
}

func ReadFile(filename string) ([]Lattice, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	return Read(file)
}

func WriteFile(filename string, sents []Lattice) error {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	Write(file, sents)
	return nil
}

func Graph2Lattice(graph NLP.LabeledDependencyGraph) Sentence {
	sent := make(Sentence, graph.NumberOfNodes()-1)
	arcIndex := make(map[int]NLP.LabeledDepArc, graph.NumberOfNodes())
	var (
		posTag string
		node   NLP.DepNode
		arc    NLP.LabeledDepArc
		headID int
		depRel string
	)
	for _, arcID := range graph.GetEdges() {
		arc = graph.GetLabeledArc(arcID)
		if arc == nil {
			panic("Can't find arc")
		}
		arcIndex[arc.GetModifier()] = arc
	}
	for _, nodeID := range graph.GetVertices() {
		if nodeID == 0 {
			continue
		}
		node = graph.GetNode(nodeID)
		posTag = ""

		taggedToken, ok := node.(*Transition.TaggedDepNode)
		if ok {
			posTag = taggedToken.POS
		}

		if node == nil {
			panic("Can't find node")
		}
		arc, exists := arcIndex[node.ID()]
		if exists {
			headID = arc.GetHead()
			depRel = string(arc.GetRelation())
		} else {
			headID = 0
			depRel = ""
		}
		row := Row{
			ID:      node.ID(),
			Form:    node.String(),
			CPosTag: posTag,
			PosTag:  posTag,
			Feats:   nil,
			Head:    headID,
			DepRel:  depRel,
		}
		sent[row.ID] = row
	}
	return sent
}

func Graph2LatticeCorpus(corpus []NLP.LabeledDependencyGraph) []Sentence {
	sentCorpus := make([]Sentence, len(corpus))
	for i, graph := range corpus {
		sentCorpus[i] = Graph2Conll(graph)
	}
	return sentCorpus
}

func Lattice2Graph(sent Sentence) NLP.LabeledDependencyGraph {
	var (
		arc  *Transition.BasicDepArc
		node NLP.DepNode
	)
	nodes := make([]NLP.DepNode, len(sent)+1)
	arcs := make([]*Transition.BasicDepArc, len(sent))
	nodes[0] = NLP.DepNode(&Transition.TaggedDepNode{0, Transition.ROOT_TOKEN, Transition.ROOT_TOKEN})
	for i, row := range sent {
		node = NLP.DepNode(&Transition.TaggedDepNode{i + 1, row.Form, row.PosTag})
		arc = &Transition.BasicDepArc{row.Head, NLP.DepRel(row.DepRel), i}
		nodes[i] = node
		arcs[i-1] = arc
	}
	return NLP.LabeledDependencyGraph(&Transition.BasicDepGraph{nodes, arcs})
}

func Lattice2GraphCorpus(corpus Lattices) []NLP.LabeledDependencyGraph {
	graphCorpus := make([]NLP.LabeledDependencyGraph, len(corpus))
	for i, sent := range corpus {
		graphCorpus[i] = Conll2Graph(sent)
	}
	return graphCorpus
}
