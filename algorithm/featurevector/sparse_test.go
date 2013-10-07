package featurevector

import (
	"testing"
)

type SparseTest struct {
	t    *testing.T
	vec1 Sparse
	vec2 Sparse
}

func (v *SparseTest) Init() {
	v.vec1, v.vec2 = make(Sparse), make(Sparse)
	v.vec1[Feature("only1")] = 1.0
	v.vec1[Feature("a")] = 1.0
	v.vec1[Feature("b")] = 0.5
	v.vec1[Feature("c")] = -0.5

	v.vec2[Feature("a")] = 1.0
	v.vec2[Feature("b")] = 2.0
	v.vec2[Feature("c")] = 0.0
	v.vec2[Feature("only2")] = 3.0
}

func (v *SparseTest) Add() {
	vec := v.vec1.Add(v.vec2)
	if vec[Feature("only1")] != 1.0 {
		v.t.Error("Got", vec[Feature("only1")], "expected", 1.0)
	}
	if vec[Feature("a")] != 2.0 {
		v.t.Error("Got", vec[Feature("a")], "expected", 2.0)
	}
	if vec[Feature("b")] != 2.5 {
		v.t.Error("Got", vec[Feature("b")], "expected", 2.5)
	}
	if vec[Feature("c")] != -0.5 {
		v.t.Error("Got", vec[Feature("c")], "expected", -0.5)
	}
	if vec[Feature("only2")] != 3.0 {
		v.t.Error("Got", vec[Feature("only2")], "expected", 3.0)
	}
}

func (v *SparseTest) Subtract() {
	vec := v.vec1.Subtract(v.vec2)
	if vec[Feature("only1")] != 1.0 {
		v.t.Error("Got", vec[Feature("only1")], "expected", 1.0)
	}
	if vec[Feature("a")] != 0.0 {
		v.t.Error("Got", vec[Feature("a")], "expected", 0.0)
	}
	if vec[Feature("b")] != -1.5 {
		v.t.Error("Got", vec[Feature("b")], "expected", -1.5)
	}
	if vec[Feature("c")] != -0.5 {
		v.t.Error("Got", vec[Feature("c")], "expected", -0.5)
	}
	if vec[Feature("only2")] != -3.0 {
		v.t.Error("Got", vec[Feature("only2")], "expected", -3.0)
	}

}

func (v *SparseTest) DotProduct() {
	dot := v.vec1.DotProduct(v.vec2)
	if dot != 2.0 {
		v.t.Error("Expected dot product", 2.0, "got", dot)
	}
}

func (v *SparseTest) FeatureWeights() {
	features := []Feature{"only1", "a", "b"}
	weights := v.vec1.FeatureWeights(features)
	if weights[Feature("only1")] != 1.0 {
		v.t.Error("Got", weights[Feature("only1")], "expected", 1.0)
	}
	if weights[Feature("a")] != 1.0 {
		v.t.Error("Got", weights[Feature("a")], "expected", 1.0)
	}
	if weights[Feature("b")] != 0.5 {
		v.t.Error("Got", weights[Feature("b")], "expected", 0.5)
	}
}

func (v *SparseTest) DotProductFeatures() {
	features := []Feature{"only1", "a", "b", "c"}
	dot := v.vec1.DotProductFeatures(features)
	if dot != 2.0 {
		v.t.Error("Expected dot product", 2.0, "got", dot)
	}
}

func (v *SparseTest) UpdateSubtract() {
	v.vec1.UpdateSubtract(v.vec2)
	if v.vec1[Feature("only1")] != 1.0 {
		v.t.Error("Got", v.vec1[Feature("only1")], "expected", 1.0)
	}
	if v.vec1[Feature("a")] != 0.0 {
		v.t.Error("Got", v.vec1[Feature("a")], "expected", 0.0)
	}
	if v.vec1[Feature("b")] != -1.5 {
		v.t.Error("Got", v.vec1[Feature("b")], "expected", -1.5)
	}
	if v.vec1[Feature("c")] != -0.5 {
		v.t.Error("Got", v.vec1[Feature("c")], "expected", -0.5)
	}
	if v.vec1[Feature("only2")] != -3.0 {
		v.t.Error("Got", v.vec1[Feature("only2")], "expected", -3.0)
	}

}

func (v *SparseTest) UpdateAdd() {
	v.vec1.UpdateAdd(v.vec2)
	if v.vec1[Feature("only1")] != 1.0 {
		v.t.Error("Got", v.vec1[Feature("only1")], "expected", 1.0)
	}
	if v.vec1[Feature("a")] != 1.0 {
		v.t.Error("Got", v.vec1[Feature("a")], "expected", 1.0)
	}
	if v.vec1[Feature("b")] != 0.5 {
		v.t.Error("Got", v.vec1[Feature("b")], "expected", 0.5)
	}
	if v.vec1[Feature("c")] != -0.5 {
		v.t.Error("Got", v.vec1[Feature("c")], "expected", -0.5)
	}
	if v.vec1[Feature("only2")] != 0.0 {
		v.t.Error("Got", v.vec1[Feature("only2")], "expected", 0.0)
	}
}

func (v *SparseTest) UpdateScalarDivide() {
	v.vec1.UpdateScalarDivide(1.0)
	if v.vec1[Feature("only1")] != 1.0 {
		v.t.Error("Got", v.vec1[Feature("only1")], "expected", 1.0)
	}
	if v.vec1[Feature("a")] != 1.0 {
		v.t.Error("Got", v.vec1[Feature("a")], "expected", 1.0)
	}
	if v.vec1[Feature("b")] != 0.5 {
		v.t.Error("Got", v.vec1[Feature("b")], "expected", 0.5)
	}
	if v.vec1[Feature("c")] != -0.5 {
		v.t.Error("Got", v.vec1[Feature("c")], "expected", -0.5)
	}
	if v.vec1[Feature("only2")] != 0.0 {
		v.t.Error("Got", v.vec1[Feature("only2")], "expected", 0.0)
	}
	v.vec1.UpdateScalarDivide(2.0)
	if v.vec1[Feature("only1")] != 0.5 {
		v.t.Error("Got", v.vec1[Feature("only1")], "expected", 0.5)
	}
	if v.vec1[Feature("a")] != 0.5 {
		v.t.Error("Got", v.vec1[Feature("a")], "expected", 0.5)
	}
	if v.vec1[Feature("b")] != 0.25 {
		v.t.Error("Got", v.vec1[Feature("b")], "expected", 0.25)
	}
	if v.vec1[Feature("c")] != -0.25 {
		v.t.Error("Got", v.vec1[Feature("c")], "expected", -0.25)
	}
	if v.vec1[Feature("only2")] != 0.0 {
		v.t.Error("Got", v.vec1[Feature("only2")], "expected", 0.0)
	}
}

func TestSparse(t *testing.T) {
	test := &SparseTest{t: t}
	test.Init()
	test.Add()
	test.Subtract()
	test.DotProduct()
	test.DotProductFeatures()
	test.FeatureWeights()
	test.UpdateSubtract()
	test.UpdateAdd()
	test.UpdateScalarDivide()
}
