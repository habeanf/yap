package FeatureVector

import "testing"

func TestHistoryValue(t *testing.T) {
	var h *HistoryValue
	// test average of single value (no integration)
	h = NewHistoryValue(0, 0.0)
	h.Increment(1)
	h.Integrate(1)
	if h.Value != 1.0 {
		t.Error("Expected 1.0 average")
	}
	// test average of same value multiple occurences

	// test average of two values same number of occurences

	// test average with ratio occurence

	// test extreme example
}
