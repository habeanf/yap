package Dependency

import (
	. "chukuparser/NLP"
)

type ConstraintModel interface{}

type ParameterModelValue interface {
	Score() float64
	ValueWith(other interface{}) ParameterModelValue
	Increment(interface{})
	Decrement(interface{})

	Copy() ParameterModelValue
}

type ParameterModel interface {
	NewModelValue() ParameterModelValue
	ModelValue(interface{}) ParameterModelValue
	ModelValueOnes(interface{}) ParameterModelValue
	Model() interface{}
	WeightedValue(ParameterModelValue) ParameterModelValue
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
