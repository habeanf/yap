package disambig

import (
	. "chukuparser/nlp/types"
	"fmt"
	"strings"
)

const (
	SEPARATOR = ";"
)

var Main_POS map[string]bool

func init() {
	Main_POS_Types := []string{"NN", "VB", "RR", "VB"}
	Main_POS = make(map[string]bool, len(Main_POS_Types))
	for _, pos := range Main_POS_Types {
		Main_POS[pos] = true
	}
}

func Full(s Spellout) string {
	return s.AsString()
}

type MProject func(m *EMorpheme) string

func projectMorphemes(s Spellout, f MProject) string {
	strs := make([]string, len(s))
	for i, morph := range s {
		strs[i] = f(morph)
	}
	return strings.Join(strs, SEPARATOR)
}

func Segments(m *EMorpheme) string {
	return m.Form
}

func POS_Props(m *EMorpheme) string {
	return fmt.Sprintf("%s_%s", m.POS, m.FeatureStr)
}

func Funcs_Main_POS_Props(m *EMorpheme) string {
	_, exists = Main_POS[morph.POS]
	if _, exists := Main_POS[morph.POS]; exists {
		return fmt.Sprintf("%s_%s", morph.POS, morph.FeatureStr)
	} else {
		return morph.Form
	}
}

func POS(m *EMorpheme) string {
	return projectMorphemes(s, func(m *EMorpheme) string {
		return m.POS
	})
}
