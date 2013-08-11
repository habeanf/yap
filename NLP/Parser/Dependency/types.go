package Dependency

import (
	. "chukuparser/NLP"
)

type ConstraintModel interface{}

type ParameterModelValue interface {
	Score(interface{}) float64
	ScoreWith(interface{}, interface{}) float64
	Increment(interface{})
	Decrement(interface{})
}

type ParameterModel interface {
	NewModelValue() ParameterModelValue
	ModelValue(interface{}) interface{}
	Model() interface{}
}

type DependencyParser interface {
	Parse(Sentence, ConstraintModel, ParameterModel) (DependencyGraph, interface{})
}

type Dependency struct {
	Constraints ConstraintModel
	Parameters  ParameterModel
	Parser      DependencyParser
}

func (d *Dependency) Parse(sent Sentence) (DependencyGraph, interface{}) {
	return d.Parser.Parse(sent, d.Constraints, d.Parameters)
}
