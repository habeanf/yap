package types

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

const (
	SEPARATOR = ";"
)

var (
	Main_POS map[string]bool
	MDParams map[string]MDParam = map[string]MDParam{
		"Form":                            Form,
		"Form_Prop":                       Form_Prop,
		"POS":                             POS,
		"POS_Prop":                        POS_Prop,
		"Form_POS_Prop":                   Form_POS_Prop,
		"Form_Lemma_POS_Prop":             Form_Lemma_POS_Prop,
		"Form_POS":                        Form_POS,
		"Funcs_Main_POS_Both_Prop":        Funcs_Main_POS_Both_Prop,
		"Funcs_Main_POS_Both_Prop_Clitic": Funcs_Main_POS_Both_Prop_Clitic,
		"Funcs_Main_POS_Both_Prop_WLemma": Funcs_Main_POS_Both_Prop_WLemma,
		"Funcs_Main_POS":                  Funcs_Main_POS,
		"Funcs_Main_POS_Prop":             Funcs_Main_POS_Prop,
	}
	AllParamFuncNames string
)

func InitOpenParamFamily(pType string) {
	var Main_POS_Types []string
	switch pType {
	case "HEBTB":
		Main_POS_Types = []string{"ADVERB", "BN", "BNT", "CD", "CDT", "JJ", "JJT", "NN", "NNP", "NNT", "RB", "VB"}
		break
	case "UD":
		Main_POS_Types = []string{"ADJ", "AUX", "ADV", "PUNCT", "NUM", "INTJ", "NOUN", "PROPN", "VERB"}
		break
	default:
		panic(fmt.Sprintf("Unknown open class family %s", pType))
	}
	log.Println("Using Family", pType, "of Main_POS_Types [", Main_POS_Types, "]")
	InitOpenParamTypes(Main_POS_Types)
}

func InitOpenParamTypes(Main_POS_Types []string) {
	Main_POS = make(map[string]bool, len(Main_POS_Types))
	for _, pos := range Main_POS_Types {
		Main_POS[pos] = true
	}
}

func init() {
	// InitOpenParamFamily("HEBTB")
	paramFuncStrs := make([]string, 0, len(MDParams))
	for k, _ := range MDParams {
		paramFuncStrs = append(paramFuncStrs, k)
	}
	sort.Strings(paramFuncStrs)
	AllParamFuncNames = strings.Join(paramFuncStrs, ", ")
}

// func Full(s Spellout) string {
// 	return s.AsString()
// }

type MDParam func(m *EMorpheme) string

func ProjectSpellout(s Spellout, f MDParam) string {
	strs := make([]string, len(s))
	for i, morph := range s {
		strs[i] = f(morph)
	}
	return strings.Join(strs, SEPARATOR)
}

func Form(m *EMorpheme) string {
	return m.Form
}

func Lemma(m *EMorpheme) string {
	return m.Lemma
}

func Form_Prop(m *EMorpheme) string {
	return fmt.Sprintf("%s_%s", m.Form, m.FeatureStr)
}

func POS(m *EMorpheme) string {
	return m.CPOS
}

func POS_Prop(m *EMorpheme) string {
	return fmt.Sprintf("%s_%s", m.CPOS, m.FeatureStr)
}

func Form_POS_Prop(m *EMorpheme) string {
	return fmt.Sprintf("%s_%s_%s", m.Form, m.CPOS, m.FeatureStr)
}

func Lemma_POS_Prop(m *EMorpheme) string {
	return fmt.Sprintf("%s_%s_%s", m.Lemma, m.CPOS, m.FeatureStr)
}

func Form_Lemma_POS_Prop(m *EMorpheme) string {
	return fmt.Sprintf("%s_%s_%s_%s", m.Form, m.Lemma, m.CPOS, m.FeatureStr)
}

func Form_POS(m *EMorpheme) string {
	return fmt.Sprintf("%s_%s", m.Form, m.CPOS)
}

func Funcs_Main_POS_No_Prop(m *EMorpheme) string {
	if _, exists := Main_POS[m.CPOS]; exists {
		return fmt.Sprintf("%s", m.CPOS)
	} else {
		return fmt.Sprintf("%s_%s_%s", m.Form, m.CPOS, m.FeatureStr)
	}
}

func Funcs_Main_POS_Both_Prop(m *EMorpheme) string {
	if _, exists := Main_POS[m.CPOS]; exists {
		return fmt.Sprintf("%s_%s", m.CPOS, m.FeatureStr)
	} else {
		return fmt.Sprintf("%s_%s_%s", m.Form, m.CPOS, m.FeatureStr)
	}
}

func Funcs_Main_POS_Both_Prop_Clitic(m *EMorpheme) string {
	if _, exists := Main_POS[m.CPOS]; exists {
		if len(m.Form) > 1 && strings.HasSuffix(m.Form, "_") {
			return fmt.Sprintf("s_%s_%s", m.CPOS, m.FeatureStr)
		} else {
			return fmt.Sprintf("%s_%s", m.CPOS, m.FeatureStr)
		}
	} else {
		return fmt.Sprintf("%s_%s_%s", m.Form, m.CPOS, m.FeatureStr)
	}
}

func Funcs_Main_POS_Both_Prop_WLemma(m *EMorpheme) string {
	return fmt.Sprintf("%s_%s_%s", m.Lemma, m.CPOS, m.FeatureStr)
	// if _, exists := Main_POS[m.CPOS]; exists {
	// 	return fmt.Sprintf("%s_%s", m.CPOS, m.FeatureStr)
	// } else {
	// 	return fmt.Sprintf("%s_%s_%s", m.Lemma, m.CPOS, m.FeatureStr)
	// }
}

func Funcs_All_WLemma(m *EMorpheme) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", m.Form, m.Lemma, m.CPOS, m.POS, m.FeatureStr)
}

func Funcs_Lemma_Main_POS(m *EMorpheme) string {
	if _, exists := Main_POS[m.CPOS]; exists {
		return fmt.Sprintf("%s", m.CPOS)
	} else {
		return fmt.Sprintf("%s_%s", m.Lemma, m.CPOS)
	}
}

func Funcs_Main_POS(m *EMorpheme) string {
	if _, exists := Main_POS[m.CPOS]; exists {
		return fmt.Sprintf("%s", m.CPOS)
	} else {
		return fmt.Sprintf("%s_%s", m.Form, m.CPOS)
	}
}

func Funcs_Main_POS_Prop(m *EMorpheme) string {
	if _, exists := Main_POS[m.CPOS]; exists {
		return fmt.Sprintf("%s_%s", m.CPOS, m.FeatureStr)
	} else {
		return fmt.Sprintf("%s_%s_%s", m.Form, m.CPOS, m.FeatureStr)
	}
}
