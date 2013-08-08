package Transition

import (
	. "chukuparser/Algorithm/Transition"
	"reflect"
	"testing"
)

var TEST_EAGER_TRANSITIONS []string = []string{
	"SH",
	"LA-ATT",
	"SH",
	"LA-SBJ",
	"RA-PRED",
	"SH",
	"LA-ATT",
	"RA-OBJ",
	"RA-ATT",
	"SH",
	"LA-ATT",
	"RA-PC",
	"RE",
	"RE",
	"RE",
	"RA-PU"}

// func TestArcEagerTransitions(t *testing.T) {
// 	conf := new(SimpleConfiguration)
// 	conf.Init(TEST_SENT)

// 	arcEag := new(ArcEager)
// 	// SHIFT
// 	shConf := arcEag.Transition(conf, Transition("SHIFT")).(*SimpleConfiguration)
// 	if qPeek, qPeekExists := shConf.Queue().Peek(); !qPeekExists || qPeek != 2 {
// 		if !qPeekExists {
// 			t.Error("Expected N0")
// 		} else {
// 			t.Error("Expected N0 = 2, got", qPeek)
// 		}
// 	}
// 	if sPeek, sPeekExists := shConf.Stack().Peek(); !sPeekExists || sPeek != 1 {
// 		if !sPeekExists {
// 			t.Error("Expected S0")
// 		} else {
// 			t.Error("Expected S0 = 1, got", sPeek)
// 		}
// 	}
// 	// LA
// 	laConf := arcEag.Transition(shConf, Transition("LA-ATT")).(*SimpleConfiguration)
// 	if sPeek, sPeekExists := laConf.Stack().Peek(); !sPeekExists || sPeek != 0 {
// 		if !sPeekExists {
// 			t.Error("Expected N0")
// 		} else {
// 			t.Error("Expected N0 = 0, got", sPeek)
// 		}
// 	}
// 	if arcs := laConf.Arcs().Get(&BasicDepArc{2, "ATT", 1}); len(arcs) != 1 {
// 		t.Error("Left arc not found, arcs: ", laConf.StringArcs())
// 	}
// 	recovered := false

// 	// LA checks conditions
// 	panicFunc := func() {
// 		defer func() {
// 			r := recover()
// 			recovered = r != nil
// 		}()
// 		_ = arcEag.Transition(laConf, Transition("LA"))
// 	}
// 	panicFunc()
// 	if !recovered {
// 		t.Error("Did not panic when trying to Left-Arc with root as stack head")
// 	}
// 	// fast forward to RA
// 	interimTransitions := TEST_STANDARD_TRANSITIONS[2:11]
// 	c := Configuration(laConf)
// 	for _, transition := range interimTransitions {
// 		if transition[:2] == "RA" {
// 			panic("Shouldn't execute untested transition")
// 		}
// 		c = arcEag.Transition(c, Transition(transition))
// 	}
// 	// RA
// 	raConf := arcEag.Transition(c, Transition("RA-PC")).(*SimpleConfiguration)
// 	if qPeek, qPeekExists := raConf.Queue().Peek(); !qPeekExists || qPeek != 6 {
// 		if !qPeekExists {
// 			t.Error("Expected N0")
// 		} else {
// 			t.Error("Expected N0 == 6, to", qPeek)
// 		}
// 	}
// 	if sPeek, sPeekExists := raConf.Stack().Peek(); !sPeekExists || sPeek == 6 {
// 		if !sPeekExists {
// 			t.Error("Expected N0")
// 		} else {
// 			t.Error("Expected N0 != 6")
// 		}
// 	}
// 	if arcs := raConf.Arcs().Get(&BasicDepArc{6, "PC", 8}); len(arcs) != 1 {
// 		t.Error("Left arc not found")
// 	}
// }

func TestArcEagerOracle(t *testing.T) {
	goldGraph := GetTestDepGraph()

	conf := Configuration(new(SimpleConfiguration))
	conf.Init(TEST_SENT)

	arcEag := new(ArcEager)
	arcEag.AddDefaultOracle()
	oracle := arcEag.Oracle()
	oracle.SetGold(goldGraph)
	for i, expected := range TEST_EAGER_TRANSITIONS {
		transition := oracle.GetTransition(conf)
		if string(transition)[:2] != expected[:2] {
			t.Error("Oracle failed at transition", i, "expected", expected, "got", transition)
		}
		conf = arcEag.Transition(conf, Transition(transition))
	}
	if !conf.Terminal() {
		t.Error("Configuration should be terminal at end of expected transition sequence")
	}

	expectedArcSet := NewArcSetSimpleFromGraph(goldGraph)
	if !expectedArcSet.Equal(conf.(*SimpleConfiguration).Arcs()) {
		t.Error("Oracle/Gold parsing resulted in wrong dependency graph")
	}
}

func TestArcEagerEsotericFunctions(t *testing.T) {
	arcEag := new(ArcEager)
	transitions := arcEag.TransitionTypes()
	if !reflect.DeepEqual(transitions, []Transition{"LA-*", "RA-*", "SH", "RE"}) {
		t.Error("Wrong transition types")
	}

}
