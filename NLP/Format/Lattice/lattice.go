package Lattice

// Package Lattice reads lattice format files

import (
	"chukuparser/Algorithm/Graph"
	NLP "chukuparser/NLP/Types"

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
	}
	return strings.Join(fields, "\t")
}

type Lattice map[int][]Edge

type Lattices []Lattice

const (
	FIELD_SEPARATOR      = '\t'
	NUM_FIELDS           = 8
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
	val := value
	if val == "_" {
		val = ""
	}
	return val
}

func ParseFeatures(featuresStr string) (Features, error) {
	var featureMap Features = make(Features)
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
		switch len(featureKV) {
		case 1:
			featureMap[featureKV[0]] = featureKV[0]
		case 2:
			featName := featureKV[0]
			featValue := featureKV[1]
			existingFeatValue, featExist := featureMap[featName]
			if featExist {
				featureMap[featName] = existingFeatValue + FEATURE_CONCAT_DELIM + featValue
			} else {
				featureMap[featName] = featValue
			}
		default:
			return nil, errors.New("Wrong number of fields for split of feature" + featureStr)
		}
	}
	return featureMap, nil
}

func ParseEdge(record []string) (*Edge, error) {
	row := &Edge{}
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
		return nil, err
	}

	var (
		currentLatt          Lattice = nil
		prevRecordFirstField string  = ""
	)
	for i, record := range records {
		// a record with id '1' indicates a new sentence
		// since csv reader ignores empty lines
		// TODO: fix to work with empty lines as new sentence indicator
		if record[0] == "0" && prevRecordFirstField != "0" {
			// store current sentence
			if currentLatt != nil {
				sentences = append(sentences, currentLatt)
			}
			currentLatt = make(Lattice)
		}
		prevRecordFirstField = record[0]

		edge, err := ParseEdge(record)
		if edge.Start == edge.End {
			continue
		}
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Error processing record %d at statement %d: %s", i, len(sentences), err.Error()))
		}
		edges, exists := currentLatt[edge.Start]
		if exists {
			currentLatt[edge.Start] = append(edges, *edge)
		} else {
			currentLatt[edge.Start] = []Edge{*edge}
		}
	}
	sentences = append(sentences, currentLatt)
	return sentences, nil
}

func Write(writer io.Writer, lattices []Lattice) error {
	for _, lattice := range lattices {
		for i := 1; i < len(lattice); i++ {
			row := lattice[i]
			for _, edge := range row {
				writer.Write(append([]byte(edge.String()), '\n'))
			}
		}
		writer.Write([]byte{'\n'})
	}
	return nil
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

func Lattice2Sentence(lattice Lattice) NLP.LatticeSentence {
	tokenSizes := make(map[int]int)
	var maxToken int = 0
	for _, edges := range lattice {
		for _, edge := range edges {
			curval, _ := tokenSizes[edge.Token]
			tokenSizes[edge.Token] = curval + 1
			if edge.Token > maxToken {
				maxToken = edge.Token
			}
		}
	}
	sent := make(NLP.LatticeSentence, maxToken+1)
	sent[0] = NLP.NewRootLattice()
	for _, edges2 := range lattice {
		for _, edge2 := range edges2 {
			lat := &sent[edge2.Token]
			if lat.Morphemes == nil {
				lat.Morphemes = make(NLP.Morphemes, 0, tokenSizes[edge2.Token])
			}
			newMorpheme := &NLP.Morpheme{
				Graph.BasicDirectedEdge{len(lat.Morphemes), edge2.Start, edge2.End},
				edge2.Word,
				edge2.CPosTag,
				edge2.PosTag,
				edge2.Feats,
				edge2.Token,
			}
			lat.Morphemes = append(lat.Morphemes, newMorpheme)
		}
	}
	for i, lat := range sent {
		lat.SortMorphemes()
		lat.GenSpellouts()
		lat.GenToken()
		sent[i] = lat
	}
	return sent
}

func Lattice2SentenceCorpus(corpus Lattices) []NLP.LatticeSentence {
	graphCorpus := make([]NLP.LatticeSentence, len(corpus))
	for i, sent := range corpus {
		graphCorpus[i] = Lattice2Sentence(sent)
	}
	return graphCorpus
}
