package lex

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	"yap/alg/graph"
	"yap/nlp/types"
)

const (
	APPROX_LEX_SIZE      = 100000
	SEPARATOR            = " "
	MSR_SEPARATOR        = ":"
	FEATURE_SEPARATOR    = "-"
	PREFIX_SEPARATOR     = "^"
	PREFIX_MSR_SEPARATOR = "+"
)

type AnalyzedToken struct {
	Token     string
	Morphemes []types.BasicMorphemes
}

func (a *AnalyzedToken) NumMorphemes() (num int) {
	for _, m := range a.Morphemes {
		num += len(m)
	}
	return
}

func ParseMSR(msr string) (string, string, map[string]string, string, error) {
	hostMSR := strings.Split(msr, FEATURE_SEPARATOR)
	return hostMSR[0], hostMSR[0], nil, strings.Join(hostMSR[1:], "|"), nil
}

func ParseMSRSuffix(msr string) (string, map[string]string, string, error) {
	hostMSR := strings.Split(msr, FEATURE_SEPARATOR)
	return "הם", nil, strings.Join(hostMSR[1:], "|"), nil
}

func ProcessAnalyzedToken(analysis string) (*AnalyzedToken, error) {
	var (
		split, msrs    []string
		curToken       *AnalyzedToken
		i              int
		curNode, curID int
		lemma          string
		def            bool
	)
	split = strings.Split(analysis, SEPARATOR)
	splitLen := len(split)
	if splitLen < 3 || splitLen%2 != 1 {
		return nil, errors.New("Wrong number of fields (" + analysis + ")")
	}
	curToken = &AnalyzedToken{
		Token:     split[0],
		Morphemes: make([]types.BasicMorphemes, 0, (splitLen-1)/2),
	}
	for i = 1; i < splitLen; i += 2 {
		curNode, curID = 0, 0
		morphs := make(types.BasicMorphemes, 0, 4)
		msrs = strings.Split(split[i], MSR_SEPARATOR)
		lemma = split[i+1]
		def = false
		// Prefix morpheme (if exists)
		if len(msrs[0]) > 0 {
			if msrs[0] == "DEF" {
				def = true
			} else {
				return nil, errors.New("Unknown prefix MSR(" + msrs[0] + ")")
			}
		}
		if len(msrs[1]) == 0 {
			return nil, errors.New("Empty host MSR (" + analysis + ")")
		}
		// Host morpheme
		CPOS, POS, Features, FeatureStr, err := ParseMSR(msrs[1])
		if err != nil {
			return nil, err
		}
		if def {
			Features["def"] = "D"
		}
		morphs = append(morphs, &types.Morpheme{
			BasicDirectedEdge: graph.BasicDirectedEdge{curID, curNode, curNode + 1},
			Form:              split[0],
			Lemma:             lemma,
			CPOS:              CPOS,
			POS:               POS,
			Features:          Features,
			TokenID:           0,
			FeatureStr:        FeatureStr,
		})
		curID++
		curNode++
		// Postfix morphemes
		if len(msrs[2]) > 0 && msrs[2][0] == 'S' {
			morphs = append(morphs, &types.Morpheme{
				BasicDirectedEdge: graph.BasicDirectedEdge{curID, curNode, curNode + 1},
				Form:              "של",
				Lemma:             "",
				CPOS:              "POS",
				POS:               "POS",
				Features:          nil,
				TokenID:           0,
				FeatureStr:        "",
			})
			curID++
			curNode++
			sufForm, sufFeatures, sufFeatureStr, err := ParseMSRSuffix(msrs[2])
			if err != nil {
				return nil, err
			}
			morphs = append(morphs, &types.Morpheme{
				BasicDirectedEdge: graph.BasicDirectedEdge{curID, curNode, curNode + 1},
				Form:              sufForm,
				Lemma:             "",
				CPOS:              "S_PRN",
				POS:               "S_PRN",
				Features:          sufFeatures,
				TokenID:           0,
				FeatureStr:        sufFeatureStr,
			})
			curID++
			curNode++
		}
		curToken.Morphemes = append(curToken.Morphemes, morphs)
	}
	return curToken, nil
}

func ProcessAnalyzedPrefix(analysis string) (*AnalyzedToken, error) {
	var (
		split, forms, prefix_msrs, msrs []string
		curToken                        *AnalyzedToken
		i                               int
		curNode, curID                  int
	)
	split = strings.Split(analysis, SEPARATOR)
	splitLen := len(split)
	if splitLen < 3 || splitLen%2 != 1 {
		return nil, errors.New("Wrong number of fields (" + analysis + ")")
	}
	curToken = &AnalyzedToken{
		Token:     split[0],
		Morphemes: make([]types.BasicMorphemes, 0, (splitLen-1)/2),
	}
	for i = 1; i < splitLen; i += 2 {
		curNode, curID = 0, 0
		morphs := make(types.BasicMorphemes, 0, 4)
		forms = strings.Split(split[i], PREFIX_SEPARATOR)
		prefix_msrs = strings.Split(split[i+1], PREFIX_MSR_SEPARATOR)
		if len(forms) != len(prefix_msrs) {
			return nil, errors.New("Mismatch between # of forms and # of MSRs (" + analysis + ")")
		}
		for j := 0; j < len(forms); j++ {
			msrs = strings.Split(prefix_msrs[j], MSR_SEPARATOR)
			// Add prefix morpheme
			if len(msrs[0]) > 0 {
				// replace -SUBCONJ for TEMP-SUBCONJ/REL-SUBCONJ
				morphs = append(morphs, &types.Morpheme{
					BasicDirectedEdge: graph.BasicDirectedEdge{curID, curNode, curNode + 1},
					Form:              forms[j],
					Lemma:             "",
					CPOS:              strings.Replace(prefix_msrs[j], "-SUBCONJ", "", -1),
					POS:               strings.Replace(prefix_msrs[j], "-SUBCONJ", "", -1),
					Features:          nil,
					TokenID:           0,
					FeatureStr:        "",
				})
				curID++
				curNode++
			}
		}
		curToken.Morphemes = append(curToken.Morphemes, morphs)
	}
	return curToken, nil
}

type LexReader func(string) (*AnalyzedToken, error)

func Read(input io.Reader, format string) ([]*AnalyzedToken, error) {
	tokens := make([]*AnalyzedToken, 0, APPROX_LEX_SIZE)
	scan := bufio.NewScanner(input)
	var reader LexReader
	switch format {
	case "lexicon":
		reader = ProcessAnalyzedToken
	case "prefix":
		reader = ProcessAnalyzedPrefix
	default:
	}
	for scan.Scan() {
		line := scan.Text()
		token, err := reader(line)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}
func ReadFile(filename string, format string) ([]*AnalyzedToken, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	return Read(file, format)
}
