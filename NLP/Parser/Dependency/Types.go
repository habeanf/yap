package Dependency

import (
	. "chukuparser/NLP"
)

type ConstraintModel interface{}

type ParameterModel interface{}

type DependencyParseFunc func(Sentence, *ConstraintModel, *ParameterModel) (*DependencyGraph, interface{})

type Dependency struct {
	Constraints *ConstraintModel
	Parameters  *ParameterModel
	ParseFunc   DependencyParseFunc
}

func (d *Dependency) Parse(sent Sentence) (*DependencyGraph, interface{}) {
	return d.ParseFunc(sent, d.Constraints, d.Parameters)
}
