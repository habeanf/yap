package ma

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"yap/alg/graph"
	"yap/nlp/format/lattice"
	"yap/nlp/format/raw"
	. "yap/nlp/types"
	"yap/util"
)

const (
	MSR_SEPARATOR = "|"
	PUNCTUATION   = ",.|?!:;-"
)

type TrainingFile struct {
	Lattice, Raw, LatMD5, RawMD5 string
}

type TokenDictionary map[string]BasicMorphemes

type MSRFreq map[string]int

type MADict struct {
	Language  string
	NumTokens int

	// for OOV
	MaxTopPOS, MaxMSRsPerPOS int
	TopPOS                   []string
	OOVMSRs                  []string
	POSMSRs                  map[string]MSRFreq

	// data
	Files []TrainingFile
	Data  TokenDictionary
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
	if m.POSMSRs == nil {
		m.POSMSRs = make(map[string]MSRFreq, 100)
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
			m.AddAnalyses(string(curToken), lat.Morphemes.Standalone())
			m.AddMSRs(lat.Morphemes.Standalone())
		}
	}
	if m.Files == nil {
		m.Files = make([]TrainingFile, 0, 1)
	}
	m.Files = append(m.Files, TrainingFile{latticeFile, rawFile, latmd5, rawmd5})

	tokensRead := len(m.Data) - m.NumTokens
	m.NumTokens = len(m.Data)
	if m.MaxTopPOS == 0 {
		m.MaxTopPOS = 5
	}
	if m.MaxMSRsPerPOS == 0 {
		m.MaxMSRsPerPOS = 10
	}

	m.computeTopPOS()
	m.computeOOVMSRs()

	return tokensRead, nil
}

func (m *MADict) computeTopPOS() {
	// compute frequency of CPOS in data dict
	posCnt := make(map[string]int, 100)
	for _, morphs := range m.Data {
		for _, morph := range morphs {
			if len(morph.CPOS) == 1 && strings.Contains(PUNCTUATION, morph.CPOS) {
				// punctuation specified as CPOS is skipped
				continue
			}
			if cnt, exists := posCnt[morph.CPOS]; exists {
				posCnt[morph.CPOS] = cnt + 1
			} else {
				posCnt[morph.CPOS] = 1
			}
		}
	}

	topN := util.GetTopNStrInt(posCnt, m.MaxTopPOS)

	m.TopPOS = make([]string, len(topN))
	for i, val := range topN {
		m.TopPOS[i] = val.S
	}
}

func (m *MADict) computeOOVMSRs() {
	if m.OOVMSRs == nil {
		m.OOVMSRs = make([]string, 0, len(m.TopPOS)*m.MaxMSRsPerPOS)
	}
	for _, pos := range m.TopPOS {
		msrfreq, exists := m.POSMSRs[pos]
		if !exists {
			fmt.Println("Top POS not found")
			fmt.Println("Top POSs:")
			fmt.Println(m.TopPOS)
			fmt.Println("Top MSRs by POS:")
			fmt.Println(m.POSMSRs)
			panic("Top POS does not have an MSR frequency entry")
		}
		topN := util.GetTopNStrInt(msrfreq, m.MaxMSRsPerPOS)
		for _, msrkv := range topN {
			m.OOVMSRs = append(m.OOVMSRs, strings.Join([]string{pos, msrkv.S}, MSR_SEPARATOR))
		}
	}
}

func (m *MADict) AddMSRs(morphs BasicMorphemes) {
	for _, morph := range morphs {
		msr := strings.Join([]string{morph.POS, morph.FeatureStr}, MSR_SEPARATOR)
		if freq, exists := m.POSMSRs[morph.CPOS]; exists {
			if cnt, msrexists := freq[msr]; msrexists {
				freq[msr] = cnt + 1
				m.POSMSRs[morph.CPOS] = freq
			}
		} else {
			freq := make(MSRFreq, 1000)
			freq[msr] = 1
			m.POSMSRs[morph.CPOS] = freq
		}
	}
}

func (m *MADict) AddAnalyses(token string, morphs BasicMorphemes) {
	if curAnalysis, exists := m.Data[token]; exists {
		// curAnalysis := _curAnalysis.(BasicMorphemes)
		// log.Println("\t\tFound, unioning")
		curAnalysis.Union(morphs)
		m.Data[token] = curAnalysis
		// log.Println("\t\tPost union")
		// log.Println("\t\t", m.Data[string(curToken)])
	} else {
		// log.Println("\t\tAdding")
		m.Data[token] = morphs
		// log.Println("\t\tPost adding")
		// log.Println("\t\t", m.Data[string(curToken)])
	}

}

// func (m *MADict) SetOOVs(oovs string) {
// 	m.OOVPOS = strings.Split(oovs, OOV_POS_SET_SEPARATOR)
// }

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

type AnalyzeStats struct {
	TotalTokens, OOVTokens    int
	UniqTokens, UniqOOVTokens map[string]int
}

func (a *AnalyzeStats) Init() {
	a.UniqTokens = make(map[string]int, 10000)
	a.UniqOOVTokens = make(map[string]int, 1000)
}

func (a *AnalyzeStats) AddToken(token string) {
	if cnt, exists := a.UniqTokens[token]; exists {
		a.UniqTokens[token] = cnt + 1
	} else {
		a.UniqTokens[token] = 1
	}
}

func (a *AnalyzeStats) AddOOVToken(token string) {
	if cnt, exists := a.UniqOOVTokens[token]; exists {
		a.UniqOOVTokens[token] = cnt + 1
	} else {
		a.UniqOOVTokens[token] = 1
	}
}

func (m *MADict) Analyze(input []string, stats *AnalyzeStats) (LatticeSentence, interface{}) {
	retval := make(LatticeSentence, len(input))
	var curNode, curID int
	for i, token := range input {
		if stats != nil {
			stats.TotalTokens++
			stats.AddToken(token)
		}
		lat := &retval[i]
		lat.Token = Token(token)
		if morphs, exists := m.Data[token]; exists {
			lat.Morphemes = make([]*EMorpheme, len(morphs))
			for j, morph := range morphs {
				lat.Morphemes[j] = &EMorpheme{
					Morpheme: Morpheme{
						graph.BasicDirectedEdge{curID, curNode, curNode + 1},
						morph.Form,
						morph.Lemma,
						morph.CPOS,
						morph.POS,
						morph.Features,
						i,
						morph.FeatureStr,
					},
				}
				curID++
			}
		} else {
			if stats != nil {
				stats.OOVTokens++
				stats.AddOOVToken(token)
			}
			// add morphemes for Out-Of-Vocabulary
			lat.Morphemes = make([]*EMorpheme, len(m.OOVMSRs))

			for j, msr := range m.OOVMSRs {
				split := strings.Split(msr, MSR_SEPARATOR)
				lat.Morphemes[j] = &EMorpheme{
					Morpheme: Morpheme{
						graph.BasicDirectedEdge{curID, curNode, curNode + 1},
						token,
						"_",
						split[0],
						split[1],
						nil,
						i,
						split[2],
					},
				}
				curID++
			}
		}
		curNode++
	}
	return retval, nil
}
