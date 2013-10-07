package conll

import (
	"strings"
	"testing"
)

func TestParseRow(t *testing.T) {
	row := strings.Split("1	EFRWT	_	CDT	CDT	gen=F|num=P	2	num	_	_",
		string(FIELD_SEPARATOR))

	parsed, err := ParseRow(row)
	if err != nil {
		t.Error(err.Error())
	}

	if parsed.ID != 1 {
		t.Errorf("Expected ID 1, got %d", parsed.ID)
	}

	if parsed.Form != "EFRWT" {
		t.Errorf("Expected FORM value EFRWT, got %s", parsed.Form)
	}

	if parsed.CPosTag != "CDT" {
		t.Errorf("Expected CPOSTAG value CDT, got %s", parsed.CPosTag)
	}

	if parsed.PosTag != "CDT" {
		t.Errorf("Expected POSTAG value CDT, got %s", parsed.PosTag)
	}

	if len(parsed.Feats) != 2 {
		t.Errorf("Expected 2 Features, got %d", len(parsed.Feats))
	}

	genFeature, genExists := parsed.Feats["gen"]
	if !genExists {
		t.Errorf("Feature gen not found")
	}
	if genFeature != "F" {
		t.Errorf("Expected F for gen, got %s", genFeature)
	}

	numFeature, numExists := parsed.Feats["num"]
	if !numExists {
		t.Errorf("Feature num not found")
	}
	if numFeature != "P" {
		t.Errorf("Expected P for num, got %s", genFeature)
	}

	if parsed.Head != 2 {
		t.Errorf("Expected HEAD value 2, got %d", parsed.Head)
	}
}

func TestParseSuccessWithoutParams(t *testing.T) {
	row := strings.Split("8	KF	_	TEMP	TEMP	_	3	ccomp	_	_",
		string(FIELD_SEPARATOR))

	_, err := ParseRow(row)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestParseRowWithRepeatingParams(t *testing.T) {
	row := strings.Split("19	PRCWPNW	_	NN	NN_S_PP	gen=M|num=S|suf_gen=F|suf_gen=M|suf_num=P|suf_per=1	18	pobj",
		string(FIELD_SEPARATOR))

	parsed, err := ParseRow(row)
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
