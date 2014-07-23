package perceptron

import "testing"

func TestPerceptron(t *testing.T) {

}

func TestTrivialStrategy(t *testing.T) {
	v := make(featurevector.Sparse)
	w := new(TrivialStrategy)
	w.Init(&v, 10)
	w.Update(&v)
	if &v != w.Finalize(&v) {
		t.Error("Should return trivial value")
	}
}

func TestAveragedStrategy(t *testing.T) {
	v := make(featurevector.Sparse)
	v[Feature("a")] = 4.0 // (1/8)
	v[Feature("b")] = 1.0
	w := new(AveragedStrategy)
	w.Init(&v, 4)
	w.Update(&v)
	w.Update(&v)
	avg := *(w.Finalize(&v))
	if avg[Feature("a")] != 1.0 {
		t.Error("Got averaged value", avg[Feature("a")], "expected", 1.0)
	}
	if avg[Feature("b")] != 0.25 {
		t.Error("Got averaged value", avg[Feature("a")], "expected", 0.25)
	}
}
