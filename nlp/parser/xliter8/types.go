package xliter8

import (
	"strings"
	"unicode/utf8"
)

type Interface interface {
	To(string) string
	From(string) string
}

const (
	hebrew          string = "אבגדהוזחטיכלמנסעפצקרשת0123456789\"%.,"
	english         string = "ABGDHWZXJIKLMNSEPCQRFT0123456789UO.,"
	extra_heb       string = "ךםןץף"
	extra_eng       string = "KMNCP"
	heb_suffix_from string = "כמנפצ"
	heb_suffix_to   string = "ךםןףץ"
)

type internalHebrew struct {
	H2E     map[rune]rune
	E2H     map[rune]rune
	HSuffix map[rune]rune
}

var (
	hebrewInstance internalHebrew

	PUNCT = map[string]string{
		":":   "yyCLN",
		",":   "yyCM",
		"-":   "yyDASH",
		".":   "yyDOT",
		"...": "yyELPS",
		"!":   "yyEXCL",
		"(":   "yyLRB",
		"?":   "yyQM",
		")":   "yyRRB",
		";":   "yySCLN",
		"\"":  "yyQUOT",
	}

	PUNCT_REV map[string]string
)

func mapH2E(input rune) rune {
	if result, exists := hebrewInstance.H2E[input]; exists {
		return result
	} else {
		return rune(-1)
	}
}

func mapE2H(input rune) rune {
	if result, exists := hebrewInstance.E2H[input]; exists {
		return result
	} else {
		return rune(-1)
	}
}

func init() {
	hebrewInstance = internalHebrew{
		make(map[rune]rune, len(hebrew)),
		make(map[rune]rune, len(english)),
		make(map[rune]rune, len(heb_suffix_to)),
	}
	eng_bytes := []byte(english)
	i := 0
	for _, heb := range hebrew {
		eng, _ := utf8.DecodeRune(eng_bytes[i : i+1])
		hebrewInstance.H2E[heb] = eng
		hebrewInstance.E2H[eng] = heb
		i++
	}
	eng_extra_bytes := []byte(extra_eng)
	i = 0
	for _, heb := range extra_heb {
		eng, _ := utf8.DecodeRune(eng_extra_bytes[i : i+1])
		hebrewInstance.H2E[heb] = eng
		i++
	}
	suffix_to_bytes := []byte(heb_suffix_to)
	for _, from_suf := range heb_suffix_from {
		to_suf, size := utf8.DecodeRune(suffix_to_bytes)
		suffix_to_bytes = suffix_to_bytes[size:]
		hebrewInstance.HSuffix[from_suf] = to_suf
	}
	PUNCT_REV = make(map[string]string, len(PUNCT))
	for k, v := range PUNCT {
		PUNCT_REV[v] = k
	}
}

type Hebrew struct{}

func (h *Hebrew) To(input string) string {
	if v, exists := PUNCT[input]; exists {
		return v
	}
	return strings.Map(mapH2E, input)
}
func (h *Hebrew) From(input string) string {
	if v, exists := PUNCT_REV[input]; exists {
		return v
	}
	retval := strings.Map(mapE2H, input)
	lastRune, size := utf8.DecodeLastRuneInString(retval)
	if newSuf, exists := hebrewInstance.HSuffix[lastRune]; exists {
		newSufStr := make([]byte, 4)
		numWritten := utf8.EncodeRune(newSufStr, newSuf)
		retval = retval[:len(retval)-size] + string(newSufStr[:numWritten])
	}
	return retval
}
