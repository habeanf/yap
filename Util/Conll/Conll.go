package Conll

// Package Conll reads ConLL format files
// For a description see http://ilk.uvt.nl/conll/#dataformat

import (
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

// A Row is a single parsed row of a conll data set
// *Not all fields are present
type Row struct {
	ID      int
	Form    string
	CPosTag string
	PosTag  string
	Feats   Features
	Head    int
	DepRel  string
}

// A Sentence is a map of Rows using their ids
type Sentence map[int]Row

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
		return row, errors.New(fmt.Sprintf("Error parsing ID field (%s): ", record[0], err.Error()))
	}
	row.ID = id

	form := ParseString(record[1])
	if form == "" {
		return row, errors.New("Empty FORM field")
	}
	row.Form = form

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
		return row, errors.New(fmt.Sprintf("Error parsing HEAD field (%s): ", record[6], err.Error()))
	}
	row.Head = head

	deprel := ParseString(record[7])
	if deprel == "" {
		return row, errors.New("Empty DEPREL field")
	}
	row.DepRel = deprel

	features, err := ParseFeatures(record[5])
	if err != nil {
		return row, errors.New(fmt.Sprintf("Error parsing FEATS field (%s): ", record[5], err.Error()))
	}
	row.Feats = features
	return row, nil
}

func Read(r io.Reader) ([]Sentence, error) {
	var sentences []Sentence
	reader := csv.NewReader(r)
	reader.Comma = FIELD_SEPARATOR
	reader.FieldsPerRecord = NUM_FIELDS

	records, err := reader.ReadAll()
	if err != nil {
		return nil, errors.New("Failure reading delimited file")
	}

	var currentSent Sentence = nil
	for i, record := range records {
		// a record with id '1' indicates a new sentence
		// since csv reader ignores empty lines
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
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Read(file)
}
