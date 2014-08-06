package lattice

// Package Lattice reads lattice format files

import (
	"chukuparser/alg/graph"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"

	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	// "log"
)

const (
	_COMPACT_AGGLUTINATED_H = true
)

var _COMPACTING_PREFIXES = map[string]bool{"B": true, "K": true, "L": true}

type Features map[string]string

func (f Features) String() string {
	if f != nil || len(f) == 0 {
		return "_"
	}
	return fmt.Sprintf("%v", map[string]string(f))
}

func (f Features) Union(other Features) {
	for k, v := range other {
		if curVal, exists := f[k]; exists {
			f[k] = curVal + "," + v
		} else {
			f[k] = v
		}
	}
}

func (f Features) Copy() Features {
	newF := make(Features, len(f))
	for k, v := range f {
		newF[k] = v
	}
	return newF
}

func (f Features) MorphHost() string {
	hostStrs := make([]string, 0, len(f))
	for name, value := range f {
		if len(name) > 2 && name[0:3] != "suf" {
			hostStrs = append(hostStrs, fmt.Sprintf("%v=%v", name, value))
		}
	}
	sort.Strings(hostStrs)
	return strings.Join(hostStrs, ",")
}

func (f Features) MorphSuffix() string {
	hostStrs := make([]string, 0, len(f))
	for name, value := range f {
		if len(name) > 2 && name[0:3] == "suf" {
			hostStrs = append(hostStrs, fmt.Sprintf("%v=%v", name, value))
		}
	}
	sort.Strings(hostStrs)
	return strings.Join(hostStrs, ",")
}

type Edge struct {
	Start   int
	End     int
	Word    string
	Lemma   string
	CPosTag string
	PosTag  string
	Feats   Features
	FeatStr string
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
		e.FeatStr,
		fmt.Sprintf("%d", e.Token),
	}
	return strings.Join(fields, "\t")
}

func (e *Edge) Copy() *Edge {
	newEdge := new(Edge)
	*newEdge = *e
	newEdge.Feats = e.Feats.Copy()
	return newEdge
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
	row.FeatStr = ParseString(record[6])
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

func Lattice2Sentence(lattice Lattice, eWord, ePOS, eWPOS, eMorphFeat, eMHost, eMSuffix *util.EnumSet) nlp.LatticeSentence {
	tokenSizes := make(map[int]int)
	var (
		maxToken                  int = 0
		origMorph, swallowedMorph *nlp.EMorpheme
		concat                    bool
	)
	for _, edges := range lattice {
		for _, edge := range edges {
			curval, _ := tokenSizes[edge.Token]
			tokenSizes[edge.Token] = curval + 1
			if edge.Token > maxToken {
				maxToken = edge.Token
			}
		}
	}
	sent := make(nlp.LatticeSentence, maxToken)
	// sent[0] = nlp.NewRootLattice()
	latticeSize := len(lattice)
	for sourceId := 0; sourceId <= latticeSize; sourceId++ {
		edges2, exists := lattice[sourceId]

		// a future sourceid may have been removed during processing
		// skip over these
		if !exists {
			// log.Println("Skipping sourceId", sourceId)
			continue
		}
		// log.Println("At sourceId", sourceId)
		for _, edge2 := range edges2 {
			concat = false
			origMorph = nil
			swallowedMorph = nil

			lat := &sent[edge2.Token-1]

			// compact agglutinated "H"
			if _COMPACT_AGGLUTINATED_H {
				if _, prefixExists := _COMPACTING_PREFIXES[edge2.Word]; prefixExists {
					if nextEdges, nextExists := lattice[edge2.End]; nextExists && nextEdges[0].Token == edge2.Token {
						if len(nextEdges) == 1 && nextEdges[0].Word == "H" {
							origMorph = &nlp.EMorpheme{
								Morpheme: nlp.Morpheme{
									graph.BasicDirectedEdge{len(lat.Morphemes), edge2.Start, edge2.End},
									edge2.Word,
									edge2.CPosTag,
									edge2.PosTag,
									edge2.Feats,
									edge2.Token,
									edge2.FeatStr,
								},
							}
							nextEdge := nextEdges[0]
							swallowedMorph = &nlp.EMorpheme{
								Morpheme: nlp.Morpheme{
									graph.BasicDirectedEdge{-1, nextEdge.Start, nextEdge.End},
									nextEdge.Word,
									nextEdge.CPosTag,
									nextEdge.PosTag,
									nextEdge.Feats,
									nextEdge.Token,
									nextEdge.FeatStr,
								},
							}
							concat = true

							// log.Println("\t", "Compacting at source id", sourceId)
							delete(lattice, edge2.End)
							edge2.End = nextEdges[0].End
							edge2.Word = edge2.Word + "_" + nextEdges[0].Word
							edge2.PosTag = edge2.PosTag + "_" + nextEdges[0].PosTag
							edge2.CPosTag = edge2.CPosTag + "_" + nextEdges[0].CPosTag
							// edge2.Feats.Union(nextEdges[0].Feats)
							edge2.FeatStr = edge2.FeatStr + "_" + nextEdges[0].FeatStr
						}
					}
				}
			}
			// log.Println("\t", "At morpheme (s,e) of token", edge2.Word, edge2.Start, edge2.End, edge2.Token)
			if lat.Morphemes == nil {
				// log.Println("\t", "Initialize new lattice")
				// initialize new lattice
				lat.Morphemes = make(nlp.Morphemes, 0, tokenSizes[edge2.Token])
				lat.Next = make(map[int][]int)
				lat.BottomId = edge2.Start
				lat.TopId = edge2.End
			} else {
				// log.Println("\t", "Update existing lattice")
				if edge2.Start < lat.BottomId {
					lat.BottomId = edge2.Start
				}
				if edge2.End > lat.TopId {
					lat.TopId = edge2.End
				}
			}
			if nextList, exists := lat.Next[sourceId]; exists {
				// log.Println("\t", "Append to next sourceId", sourceId)
				lat.Next[sourceId] = append(nextList, len(lat.Morphemes))
				// recheck, _ := lat.Next[sourceId]
				// log.Println("\t", "Post append:", recheck)
			} else {
				// log.Println("\t", "Create new next for sourceId", sourceId)
				lat.Next[sourceId] = make([]int, 1)
				lat.Next[sourceId][0] = len(lat.Morphemes)
			}
			newMorpheme := &nlp.EMorpheme{
				Morpheme: nlp.Morpheme{
					graph.BasicDirectedEdge{len(lat.Morphemes), edge2.Start, edge2.End},
					edge2.Word,
					edge2.CPosTag,
					edge2.PosTag,
					edge2.Feats,
					edge2.Token,
					edge2.FeatStr,
				},
				OrigMorph:      origMorph,
				SwallowedMorph: swallowedMorph,
				Concat:         concat,
			}
			newMorpheme.EForm, _ = eWord.Add(edge2.Word)
			newMorpheme.EPOS, _ = eWord.Add(edge2.CPosTag)
			newMorpheme.EFCPOS, _ = eWord.Add([2]string{edge2.Word, edge2.CPosTag})
			newMorpheme.EFeatures, _ = eMorphFeat.Add(edge2.FeatStr)
			newMorpheme.EMHost, _ = eMHost.Add(edge2.Feats.MorphHost())
			newMorpheme.EMSuffix, _ = eMSuffix.Add(edge2.Feats.MorphSuffix())
			// log.Println("\t", "Adding morpheme", newMorpheme)
			lat.Morphemes = append(lat.Morphemes, newMorpheme)
		}
	}
	for i, lat := range sent {
		// log.Println("At lat", i)
		lat.SortMorphemes()
		lat.SortNexts()
		lat.GenSpellouts()
		lat.GenToken()
		sent[i] = lat
	}
	return sent
}

func Lattice2SentenceCorpus(corpus Lattices, eWord, ePOS, eWPOS, eMorphFeat, eMHost, eMSuffix *util.EnumSet) []interface{} {
	graphCorpus := make([]interface{}, len(corpus))
	for i, sent := range corpus {
		// log.Println("At sent", i)
		graphCorpus[i] = Lattice2Sentence(sent, eWord, ePOS, eWPOS, eMorphFeat, eMHost, eMSuffix)
	}
	return graphCorpus
}
