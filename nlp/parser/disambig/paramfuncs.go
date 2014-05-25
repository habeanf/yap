package disambig

import (
	. "chukuparser/nlp/types"
	"fmt"
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

type MProject func(m Morpheme) string

func projectMorphemes(s Spellout, f MProject) string {
	strs := make([]string, len(s))
	for i, morph := range s {
		strs[i] = MProject(morph)
	}
	return strings.Join(strs, SEPARATOR)
}

func Segments(s Spellout) string {
	return projectMorphemes(s, func(m Morpheme) {
		return m.Form
	})
}

func POS_Props(s Spellout) string {
	return projectMorphemes(s, func(m Morpheme) {
		return fmt.Sprintf("%s_%s", m.POS, m.FeatureStr)
	})
}

func Funcs_Main_POS_Props(s Spellout) string {
	strs := make([]string, len(s))
	var exists bool
	for i, morph := range s {
		_, exists = Main_POS[morph.POS]
		if exists {
			strs[i] = fmt.Sprintf("%s_%s", m.POS, m.FeatureStr)
		} else {
			strs[i] = morph.Form
		}
	}
	return strings.Join(strs, SEPARATOR)
}

func POS(s Spellout) string {
	return projectMorphemes(s, func(m Morpheme) {
		return m.POS
	})
}
