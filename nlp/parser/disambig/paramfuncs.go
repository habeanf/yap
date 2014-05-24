package disambig

import (
	. "chukuparser/nlp/types"
	"strings"
)

func Full(s Spellout) string {
	morphemesAsStr := make([]string, len(s))
	for i, morph := range s {
		morphemesAsStr[i] = morph.String()
	}
	return strings.Join(morphemesAsStr, ",")
}
