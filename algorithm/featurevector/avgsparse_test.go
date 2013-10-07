package featurevector

import "testing"

func TestHistoryValue(t *testing.T) {
	var h *HistoryValue
	// test average of single occurence (integration of 1)
	h = NewHistoryValue(0, 1.0)
	h.Integrate(1)
	if h.Value != 1.0 {
		t.Errorf("Expected 1.0 average, got %v", h.Value)
	}

	// should be 0
	h = NewHistoryValue(0, 0.0)
	// value of 4, generation 2
	h.Increment(2)
	h.Increment(2)
	h.Increment(2)
	h.Increment(2)
	// value of 4 remains, integrate generation 4
	// should be average of 2
	h.Integrate(2)
	// test average of two values same number of occurences
	if h.Value != 0.0 {
		t.Errorf("Expected 0.0 average, got %v", h.Value)
	}

	// test average of same value multiple occurences
	h = NewHistoryValue(0, 0.0)
	// value of 4, generation 2
	h.Increment(2)
	h.Increment(2)
	h.Increment(2)
	h.Increment(2)
	// value of 4 remains, integrate generation 4
	// should be average of 2
	h.Integrate(4)
	// test average of two values same number of occurences
	if h.Value != 2.0 {
		t.Errorf("Expected 2.0 average, got %v", h.Value)
	}

	// test average with ratio occurence:
	// [0 x4, 20 x2, 40 x2] = 15
	h = NewHistoryValue(0, 0.0)
	h.Increment(4)
	// shortcut to set value
	h.Value = 20.0
	h.Increment(6)
	h.Value = 40.0
	h.Integrate(8)

	if h.Value != 15.0 {
		t.Errorf("Expected 15.0 average, got %v", h.Value)
	}

	// test various
}
