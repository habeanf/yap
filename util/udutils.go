package util

import (
	"fmt"
	"sort"
	"strings"
)

type FeatureLookup struct {
	UDName   string
	ValueMap map[string]string
}

var (
	GenderMap = FeatureLookup{
		UDName: "Gender",
		ValueMap: map[string]string{
			"F": "Fem",
			"M": "Masc",
		},
	}
	NumberMap = FeatureLookup{
		UDName: "Number",
		ValueMap: map[string]string{
			"S":              "Sing",
			"P":              "Plur",
			"D":              "Dual",
			"Underspecified": "Underspecified",
		},
	}
	PersonMap = FeatureLookup{
		UDName: "Person",
		ValueMap: map[string]string{
			"1": "1",
			"2": "2",
			"3": "3",
			"A": "1,2,3",
		},
	}
	DefMap = FeatureLookup{
		UDName: "Definite",
		ValueMap: map[string]string{
			"D": "Def",
			"-": "Ind",
		},
	}
	PolarMap = FeatureLookup{
		UDName: "Polarity",
		ValueMap: map[string]string{
			"pos": "Pos",
			"neg": "Neg",
		},
	}
	TenseMap = FeatureLookup{
		UDName: "Tense",
		ValueMap: map[string]string{
			"PAST":       "Past",
			"PRESENT":    "Pres",
			"FUTURE":     "Fut",
			"IMPERATIVE": "Imp", //  but will be overriden by Mood=Imp
			// note there is should be special handler for BEINONI,
			// depending on the POS
			"BEINONI": "Pres",
		},
	}
	TypeMap = FeatureLookup{
		UDName: "PronType",
		ValueMap: map[string]string{
			"DEM":  "Dem",
			"IMP":  "Ind",
			"PERS": "Prs",
			"REF":  "Prs", // additional Reflex=Yes added in code
		},
	}
	HEB2UDFeatureNameLookup = map[string]FeatureLookup{
		"gen":   GenderMap,
		"num":   NumberMap,
		"per":   PersonMap,
		"def":   DefMap,
		"tense": TenseMap,
		"type":  TypeMap,
		"polar": PolarMap,
	}
	HEB2UDPrefixPOS = map[string]string{
		"ADVERB":      "ADP",
		"CONJ":        "CCONJ",
		"DEF":         "DET",
		"PREPOSITION": "ADP",
		"REL":         "SCONJ",
		"TEMP":        "SCONJ",
	}
	HEB2UDPOS = map[string]string{
		"AT":       "PART-Case=Acc",
		"BN":       "VERB-VerbForm=Part",
		"BNT":      "VERB-Definite=Cons|VerbForm=Part",
		"CC":       "CCONJ",
		"CC-SUB":   "SCONJ",
		"CC-COORD": "SCONJ",
		"CC-REL":   "SCONJ",
		"CD":       "NUM",
		"CDT":      "NUM-Definite=Cons",
		"COP":      "AUX-VerbType=Cop",
		"DT":       "DET-Definite=Cons",
		"DTT":      "DET-Definite=Cons",
		"EX":       "VERB-HebExistential=True",
		"IN":       "ADP",
		"INTJ":     "INTJ",
		"JJ":       "ADJ",
		"JJT":      "ADJ-Definite=Cons",
		"MD":       "AUX-VerbType=Mod",
		"NEG":      "ADV",
		"NN":       "NOUN",
		"NNP":      "PROPN",
		"NNT":      "NOUN-Definite=Cons",
		"NNPT":     "PROPN-Abbvr=Yes",
		"P":        "ADV-Prefix=Yes",
		"POS":      "PART-Case=Gen",
		"PRP":      "PRON",
		"QW":       "ADV-PronType=Int",
		"RB":       "ADV-Polarity=Neg",
		"TTL":      "NOUN-Title=Yes",
		"VB":       "VERB",
		// "UNK" should be dropped
	}
)

func Heb2UDFeature(feature string) string {
	if len(feature) == 0 {
		return feature
	}
	switch feature {
	case "tense=BEINONI":
		return "Tense=Part"
	case "type=TOINFINITIVE":
		return "VerbForm=Inf"
	case "tense=IMPERATIVE":
		return "Mood=Imp"
	}
	pair := strings.Split(feature, "=")
	if len(pair) == 1 {
		panic(fmt.Sprintf("Can't transform non-attribute feature %s", feature))
	}
	if pair[0] == "binyan" {
		if pair[1] == "HITPAEL" {
			return ""
		}
		return fmt.Sprintf("HebBinyan=%s", pair[1])
	}
	if propMap, exists := HEB2UDFeatureNameLookup[pair[0]]; exists {
		if propValue, valExists := propMap.ValueMap[pair[1]]; valExists {
			return fmt.Sprintf("%s=%s", propMap.UDName, propValue)
		} else {
			panic(fmt.Sprintf("Morphological feature value does not exist in Heb2UD transform %s", feature))
		}
	} else {
		panic(fmt.Sprintf("Failed transforming feature %s", feature))
	}
}
func Heb2UDFeaturesString(features string) string {
	if features == "_" {
		return features
	}

	pairs := strings.Split(features, "|")
	udPairs := make([]string, 0, len(pairs))

	var udFeature string
	for _, hebFeature := range pairs {
		udFeature = Heb2UDFeature(hebFeature)
		if len(udFeature) > 0 {
			udPairs = append(udPairs, Heb2UDFeature(hebFeature))
		}
	}

	sort.Strings(udPairs)
	return strings.Join(udPairs, "|")
}

func MergeFeatureStrs(feat1, feat2 string) (string, map[string]string) {
	if len(feat2) > 0 && len(feat1) == 0 {
		feat1, feat2 = feat2, feat1
	}
	pair1 := strings.Split(feat1, "|")
	if len(feat2) > 0 {
		pair2 := strings.Split(feat2, "|")
		pair1 = append(pair1, pair2...)
		sort.Strings(pair1)
		pair1 = UniqueSortedStrings(pair1)
	}
	returnPairs := make([]string, 0, len(pair1))
	featureMap := make(map[string]string, len(pair1))
	for _, feature := range pair1 {
		if len(feature) > 0 {
			returnPairs = append(returnPairs, feature)
			splitFeature := strings.Split(feature, "=")
			featureMap[splitFeature[0]] = splitFeature[1]
		}
	}
	return strings.Join(returnPairs, "|"), featureMap
}

func UniqueSortedStrings(sorted []string) []string {
	if len(sorted) < 2 {
		return sorted
	}
	var prev string
	uniqueStrings := make([]string, 1, len(sorted))
	for i, cur := range sorted {
		if len(cur) == 0 {
			continue
		}
		if i > 0 && prev == cur {
			continue
		} else {
			uniqueStrings = append(uniqueStrings, cur)
			prev = cur
		}
	}
	return uniqueStrings
}

func AddToFeatureStr(featureStr, newFeature string) string {
	if len(featureStr) > 0 {
		return fmt.Sprintf("%s|%s", featureStr, newFeature)
	} else {
		return newFeature
	}
}

func DelFromFeatureMapAndStr(features map[string]string, featureStr, delFeature string) (map[string]string, string) {
	if val, exists := features[delFeature]; exists {
		delete(features, delFeature)
		if strings.Contains(featureStr, "|"+delFeature) {
			return features, strings.Replace(featureStr, "|"+delFeature+"="+val, "", -1)
		} else {
			return features, strings.Replace(featureStr, delFeature+"="+val, "", -1)
		}
	} else {
		return features, featureStr
	}
}
