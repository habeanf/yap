package ma

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"yap/nlp/format/lattice"
	"yap/nlp/format/raw"
	. "yap/nlp/types"
	"yap/util"
)

type TrainingFile struct {
	Lattice, Raw, LatMD5, RawMD5 string
}

type TokenDictionary map[string]BasicMorphemes

type MADict struct {
	Language  string
	NumTokens int
	Files     []TrainingFile
	Data      TokenDictionary
}

func (m *MADict) LearnFrom(latticeFile, rawFile string) (int, error) {
	latmd5, err := util.MD5File(latticeFile)
	if err != nil {
		return 0, err
	}
	rawmd5, err := util.MD5File(rawFile)
	if err != nil {
		return 0, err
	}
	lattices, err := lattice.ReadFile(latticeFile)
	if err != nil {
		log.Println("Error reading lattice file")
		return 0, err
	}
	tokens, err := raw.ReadFile(rawFile)
	if err != nil {
		log.Println("Error reading raw file")
		return 0, err
	}
	if len(lattices) != len(tokens) {
		log.Println("Read", len(lattices), "lattices and", len(tokens), "raw tokens")
		return 0, errors.New("Number of read sentences differ for lattice and raw files")
	}
	if m.Data == nil {
		m.Data = make(TokenDictionary)
	}
	eWord := util.NewEnumSet(100)
	ePOS := util.NewEnumSet(100)
	eWPOS := util.NewEnumSet(100)
	eMorphFeat := util.NewEnumSet(100)
	eMHost := util.NewEnumSet(100)
	eMSuffix := util.NewEnumSet(100)
	corpus := lattice.Lattice2SentenceCorpus(lattices, eWord, ePOS, eWPOS, eMorphFeat, eMHost, eMSuffix)
	for i, _sentLat := range corpus {
		// log.Println("At sentence", i)
		sentLat := _sentLat.(LatticeSentence)
		curTokens := tokens[i]
		for j, lat := range sentLat {
			curToken := curTokens[j]
			// log.Println("\tAt token", curToken)
			if curAnalysis, exists := m.Data[string(curToken)]; exists {
				// curAnalysis := _curAnalysis.(BasicMorphemes)
				// log.Println("\t\tFound, unioning")
				curAnalysis.Union(lat.Morphemes.Standalone())
				m.Data[string(curToken)] = curAnalysis
				// log.Println("\t\tPost union")
				// log.Println("\t\t", m.Data[string(curToken)])
			} else {
				// log.Println("\t\tAdding")
				m.Data[string(curToken)] = lat.Morphemes.Standalone()
				// log.Println("\t\tPost adding")
				// log.Println("\t\t", m.Data[string(curToken)])
			}
		}
	}
	if m.Files == nil {
		m.Files = make([]TrainingFile, 0, 1)
	}
	m.Files = append(m.Files, TrainingFile{latticeFile, rawFile, latmd5, rawmd5})

	tokensRead := len(m.Data) - m.NumTokens
	m.NumTokens = len(m.Data)
	return tokensRead, nil
}

func (m *MADict) Write(writer io.Writer) error {
	enc := json.NewEncoder(writer)
	err := enc.Encode(m)
	return err
}

func (m *MADict) Read(r io.Reader) error {
	dec := json.NewDecoder(r)
	err := dec.Decode(m)
	return err
}

func (m *MADict) WriteFile(filename string) error {
	file, err := os.Create(filename)
	defer file.Close()

	if err != nil {
		return err
	}

	return m.Write(file)
}

func (m *MADict) ReadFile(filename string) error {
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		return err
	}

	return m.Read(file)
}
