package Conll

// Package Conll reads ConLL format files
// For a description see http://ilk.uvt.nl/conll/#dataformat

import (
	"chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	FIELD_SEPARATOR      = '\t'
	NUM_FIELDS           = 10
	FEATURES_SEPARATOR   = "|"
	FEATURE_SEPARATOR    = "="
	FEATURE_CONCAT_DELIM = ","
)

type Features map[string]string

func (f Features) String() string {
	if f != nil || len(f) == 0 {
		return "_"
	}
	return fmt.Sprintf("%v", map[string]string(f))
}

// A Row is a single parsed row of a conll data set
// *Commented fields are not in use
type Row struct {
	ID      int
	Form    string
	CPosTag string
	PosTag  string
	Feats   Features
	Head    int
	DepRel  string
	// Lemma string
	// PHead int
	// PDepRel string
}

func (r Row) String() string {
	fields := []string{
		fmt.Sprintf("%d", r.ID),
		r.Form,
		"_",
		r.CPosTag,
		r.PosTag,
		r.Feats.String(),
		fmt.Sprintf("%d", r.Head),
		r.DepRel,
		"_",
		"_"}
	return strings.Join(fields, "\t")
}

// A Sentence is a map of Rows using their ids
type Sentence map[int]Row

type Sentences []Sentence

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

func ParseRow(record []string) (Row, error) {
	var row Row
	id, err := ParseInt(record[0])
	if err != nil {
		return row, errors.New(fmt.Sprintf("Error parsing ID field (%s): %s", record[0], err.Error()))
	}
	row.ID = id

	form := ParseString(record[1])
	if form == "" {
		return row, errors.New("Empty FORM field")
	}
	row.Form = form

	// lemma := ParseString(record[2])
	// if lemma == "" {
	// 	return row, errors.New("Empty LEMMA field")
	// }
	// row.Lemma = lemma

	cpostag := ParseString(record[3])
	if cpostag == "" {
		return row, errors.New("Empty CPOSTAG field")
	}
	row.CPosTag = cpostag

	postag := ParseString(record[4])
	if postag == "" {
		return row, errors.New("Empty POSTAG field")
	}
	row.PosTag = postag

	head, err := ParseInt(record[6])
	if err != nil {
		return row, errors.New(fmt.Sprintf("Error parsing HEAD field (%s): %s", record[6], err.Error()))
	}
	row.Head = head

	deprel := ParseString(record[7])
	if deprel == "" {
		return row, errors.New("Empty DEPREL field")
	}
	row.DepRel = deprel

	// phead, err := ParseInt(record[8])
	// if err != nil {
	// 	return row, errors.New(fmt.Sprintf("Error parsing PHEAD field (%s): %s", record[8], err.Error()))
	// }
	// row.PHead = phead

	// pdeprel := ParseString(record[9])
	// if pdeprel == "" {
	// 	return row, errors.New("Empty PDEPREL field")
	// }
	// row.PDepRel = pdeprel

	features, err := ParseFeatures(record[5])
	if err != nil {
		return row, errors.New(fmt.Sprintf("Error parsing FEATS field (%s): %s", record[5], err.Error()))
	}
	row.Feats = features
	return row, nil
}

func Read(reader io.Reader) (Sentences, error) {
	var sentences []Sentence
	csvReader := csv.NewReader(reader)
	csvReader.Comma = FIELD_SEPARATOR
	csvReader.FieldsPerRecord = NUM_FIELDS

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, errors.New("Failure reading delimited file")
	}

	var currentSent Sentence = nil
	for i, record := range records {
		// a record with id '1' indicates a new sentence
		// since csv csvReader ignores empty lines
		if record[0] == "1" {
			// store current sentence
			if currentSent != nil {
				sentences = append(sentences, currentSent)
			}
			currentSent = make(Sentence)
		}

		row, err := ParseRow(record)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Error processing record %d at statement %d: %s", i, len(sentences), err.Error()))
		}
		currentSent[row.ID] = row
	}
	sentences = append(sentences, currentSent)
	return sentences, nil
}

func ReadFile(filename string) ([]Sentence, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	return Read(file)
}

func Write(writer io.Writer, sents []Sentence) {
	for _, sent := range sents {
		for i := 1; i <= len(sent); i++ {
			row := sent[i]
			writer.Write(append([]byte(row.String()), '\n'))
		}
		writer.Write([]byte{'\n'})
	}
}

func WriteFile(filename string, sents []Sentence) error {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	Write(file, sents)
	return nil
}

func Graph2Conll(graph NLP.LabeledDependencyGraph) Sentence {
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

func Graph2ConllCorpus(corpus []NLP.LabeledDependencyGraph) []Sentence {
	sentCorpus := make([]Sentence, len(corpus))
	for i, graph := range corpus {
		sentCorpus[i] = Graph2Conll(graph)
	}
	return sentCorpus
}

func Conll2Graph(sent Sentence) NLP.LabeledDependencyGraph {
	var (
		arc  *Transition.BasicDepArc
		node NLP.DepNode
	)
	nodes := make([]NLP.DepNode, len(sent)+1)
	arcs := make([]*Transition.BasicDepArc, len(sent))
	nodes[0] = NLP.DepNode(&Transition.TaggedDepNode{0, NLP.ROOT_TOKEN, NLP.ROOT_TOKEN})
	for i, row := range sent {
		node = NLP.DepNode(&Transition.TaggedDepNode{i + 1, row.Form, row.PosTag})
		arc = &Transition.BasicDepArc{row.Head, NLP.DepRel(row.DepRel), i}
		nodes[i] = node
		arcs[i-1] = arc
	}
	return NLP.LabeledDependencyGraph(&Transition.BasicDepGraph{nodes, arcs})
}

func Conll2GraphCorpus(corpus []Sentence) []NLP.LabeledDependencyGraph {
	graphCorpus := make([]NLP.LabeledDependencyGraph, len(corpus))
	for i, sent := range corpus {
		graphCorpus[i] = Conll2Graph(sent)
	}
	return graphCorpus
}

func MorphGraph2Conll(graph NLP.MorphDependencyGraph) Sentence {
	sent := make(Sentence, graph.NumberOfNodes()-1)
	arcIndex := make(map[int]NLP.LabeledDepArc, graph.NumberOfNodes())
	var (
		node   *NLP.Morpheme
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
		node = graph.GetMorpheme(nodeID)

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
			Form:    node.Form,
			CPosTag: node.CPOS,
			PosTag:  node.POS,
			Feats:   node.Features,
			Head:    headID,
			DepRel:  depRel,
		}
		sent[row.ID] = row
	}
	return sent
}

func MorphGraph2ConllCorpus(corpus []NLP.MorphDependencyGraph) []Sentence {
	sentCorpus := make([]Sentence, len(corpus))
	for i, graph := range corpus {
		sentCorpus[i] = MorphGraph2Conll(graph)
	}
	return sentCorpus
}
