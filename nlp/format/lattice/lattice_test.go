package lattice

import (
	"strings"
	"testing"
)

func TestParseEdgeWithParams(t *testing.T) {
	row := strings.Split("0	1	EFRWT	_	CDT	CDT	gen=F|num=P	1",
		string(FIELD_SEPARATOR))

	_, err := ParseEdge(row)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestParseEdgeWithoutParams(t *testing.T) {
	row := strings.Split("4	5	TAILND	_	NNP	NNP	_	4",
		string(FIELD_SEPARATOR))

	_, err := ParseEdge(row)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestParseEdgeWithRepeatingParams(t *testing.T) {
	row := strings.Split("18	19	PRCWPNW	_	NN	NN_S_PP	gen=M|num=S|suf_gen=F|suf_gen=M|suf_num=P|suf_per=1	17",
		string(FIELD_SEPARATOR))

	parsed, err := ParseEdge(row)
	if err != nil {
		t.Error(err.Error())
	}
	value, exists := parsed.Feats["suf_gen"]
	if !exists {
		t.Error("Feature suf_gen not found")
	}
	if value != "F,M" {
		t.Error("Failure concatenating multiple features: should be F,M got " + parsed.Feats["suf_gen"])
	}
}
