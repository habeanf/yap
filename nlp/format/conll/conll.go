package conll

// Package Conll reads ConLL format files
// For a description see http://ilk.uvt.nl/conll/#dataformat

import (
	"chukuparser/nlp/parser/dependency/transition"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	// "log"
	"os"
	"sort"
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

func (f Features) MorphHost() string {
	hostStrs := make([]string, 0, len(f))
	for name, value := range f {
		if name[0:3] != "suf" {
			hostStrs = append(hostStrs, fmt.Sprintf("%v=%v", name, value))
		}
	}
	sort.Strings(hostStrs)
	return strings.Join(hostStrs, ",")
}

func (f Features) MorphSuffix() string {
	hostStrs := make([]string, 0, len(f))
	for name, value := range f {
		if name[0:3] == "suf" {
			hostStrs = append(hostStrs, fmt.Sprintf("%v=%v", name, value))
		}
	}
	sort.Strings(hostStrs)
	return strings.Join(hostStrs, ",")
}

func FormatFeatures(feat map[string]string) string {
	if feat == nil || len(feat) == 0 {
		return "_"
	}
	strs := make([]string, 0, len(feat))
	for k, v := range feat {
		strs = append(strs, fmt.Sprintf("%v%v%v", k, FEATURE_SEPARATOR, v))
	}
	sort.Strings(strs)
	return strings.Join(strs, FEATURES_SEPARATOR)
}

// A Row is a single parsed row of a conll data set
// *Commented fields are not in use
type Row struct {
	ID      int
	Form    string
	CPosTag string
	PosTag  string
	Feats   Features
	FeatStr string
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
		FormatFeatures(r.Feats),
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
	row.FeatStr = ParseString(record[5])
	return row, nil
}

func Read(reader io.Reader) (Sentences, error) {
	var sentences []Sentence
	csvReader := csv.NewReader(reader)
	csvReader.Comma = FIELD_SEPARATOR
	csvReader.FieldsPerRecord = NUM_FIELDS

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failure reading delimited file: %s", err.Error()))
	}

	var currentSent Sentence = nil
	for i, record := range records {
		// log.Println("At record", i)
		// a record with id '1' indicates a new sentence
		// since csv csvReader ignores empty lines
		if record[0] == "1" {
			// log.Println("At sentence", len(sentences))
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

func Graph2Conll(graph nlp.LabeledDependencyGraph) Sentence {
	sent := make(Sentence, graph.NumberOfNodes())
	arcIndex := make(map[int]nlp.LabeledDepArc, graph.NumberOfNodes())
	var (
		posTag string
		node   nlp.DepNode
		arc    nlp.LabeledDepArc
		headID int
		depRel string
	)
	for _, arcID := range graph.GetEdges() {
		arc = graph.GetLabeledArc(arcID)
		if arc == nil {
			// panic("Can't find arc")
		} else {
			arcIndex[arc.GetModifier()] = arc
		}
	}
	for _, nodeID := range graph.GetVertices() {
		node = graph.GetNode(nodeID)
		posTag = ""

		taggedToken, ok := node.(*transition.TaggedDepNode)
		if ok {
			posTag = taggedToken.RawPOS
		}

		if node == nil {
			panic("Can't find node")
		}
		arc, exists := arcIndex[node.ID()]
		if exists {
			headID = arc.GetHead()
			depRel = string(arc.GetRelation())
			if depRel == nlp.ROOT_LABEL {
				headID = -1
			}
		} else {
			headID = -1
			depRel = "None"
		}
		row := Row{
			ID:      node.ID() + 1,
			Form:    node.String(),
			CPosTag: posTag,
			PosTag:  posTag,
			Feats:   nil,
			Head:    headID + 1,
			DepRel:  depRel,
		}
		sent[row.ID] = row
	}
	return sent
}

func Graph2ConllCorpus(corpus []interface{}) []Sentence {
	sentCorpus := make([]Sentence, len(corpus))
	for i, graph := range corpus {
		sentCorpus[i] = Graph2Conll(graph.(nlp.LabeledDependencyGraph))
	}
	return sentCorpus
}

func Conll2Graph(sent Sentence, eWord, ePOS, eWPOS, eRel, eMHost, eMSuffix *util.EnumSet) nlp.LabeledDependencyGraph {
	var (
		arc   *transition.BasicDepArc
		node  *transition.TaggedDepNode
		index int
	)
	nodes := make([]nlp.DepNode, 0, len(sent)+2)
	// log.Println("\tNum Nodes:", len(nodes))
	arcs := make([]*transition.BasicDepArc, len(sent))
	// node.Token, _ = eWord.Add(nlp.ROOT_TOKEN)
	// node.POS, _ = ePOS.Add(nlp.ROOT_TOKEN)
	// node.TokenPOS, _ = eWPOS.Add([2]string{nlp.ROOT_TOKEN, nlp.ROOT_TOKEN})
	// nodes = append(nodes, nlp.DepNode(node)) // add root node

	for i := 1; i <= len(sent); i++ {
		row, _ := sent[i]
		// for i, row := range sent {
		node = &transition.TaggedDepNode{
			Id:       i - 1,
			RawToken: row.Form,
			RawPOS:   row.CPosTag,
		}
		node.Token, _ = eWord.Add(row.Form)
		node.POS, _ = ePOS.Add(row.CPosTag)
		node.TokenPOS, _ = eWPOS.Add([2]string{row.Form, row.CPosTag})
		node.MHost, _ = eMHost.Add(row.Feats.MorphHost())
		node.MSuffix, _ = eMSuffix.Add(row.Feats.MorphSuffix())
		index, _ = eRel.IndexOf(nlp.DepRel(row.DepRel))
		arc = &transition.BasicDepArc{row.Head - 1, index, i - 1, nlp.DepRel(row.DepRel)}
		// log.Println("Adding node", node, node.TokenPOS, eWPOS.ValueOf(node.TokenPOS))
		nodes = append(nodes, nlp.DepNode(node))
		// log.Println("Adding arc", i-1, arc)
		arcs[i-1] = arc
	}
	return nlp.LabeledDependencyGraph(&transition.BasicDepGraph{nodes, arcs})
}

func Conll2GraphCorpus(corpus []Sentence, eWord, ePOS, eWPOS, eRel, eMHost, eMSuffix *util.EnumSet) []interface{} {
	graphCorpus := make([]interface{}, len(corpus))
	for i, sent := range corpus {
		// log.Println("Converting sentence", i)
		graphCorpus[i] = Conll2Graph(sent, eWord, ePOS, eWPOS, eRel, eMHost, eMSuffix)
	}
	return graphCorpus
}

func MorphGraph2Conll(graph nlp.MorphDependencyGraph) Sentence {
	sent := make(Sentence, graph.NumberOfNodes())
	arcIndex := make(map[int]nlp.LabeledDepArc, graph.NumberOfNodes())
	var (
		node   *nlp.EMorpheme
		arc    nlp.LabeledDepArc
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
	for i, nodeID := range graph.GetVertices() {
		node = graph.GetMorpheme(nodeID)

		if node == nil {
			panic("Can't find node")
		}

		arc, exists := arcIndex[i]
		if exists {
			headID = arc.GetHead()
			depRel = string(arc.GetRelation())
			if depRel == nlp.ROOT_LABEL {
				headID = -1
			}
		} else {
			headID = 0
			depRel = "None"
		}
		row := Row{
			ID:      i + 1,
			Form:    node.Form,
			CPosTag: node.CPOS,
			PosTag:  node.POS,
			Feats:   node.Features,
			Head:    headID + 1,
			DepRel:  depRel,
		}
		sent[row.ID] = row
	}
	return sent
}

func MorphGraph2ConllCorpus(corpus []nlp.MorphDependencyGraph) []Sentence {
	sentCorpus := make([]Sentence, len(corpus))
	for i, graph := range corpus {
		sentCorpus[i] = MorphGraph2Conll(graph)
	}
	return sentCorpus
}
