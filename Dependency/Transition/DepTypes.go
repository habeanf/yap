package Transition

type ConstraintModel interface{}

type ParameterModel interface{}

type DependencyParserFunc func(*Sentence, *ConstraintModel, *ParameterModel) (*Graph, interface{})

type Dependency struct {
	Constraints *ConstraintModel
	Parameters  *ParameterModel
	ParseFunc   DependencyParseFunc
}

func (d *Dependency) Parse(sent Sentence) (*Graph, []Configuration) {
	return d.ParseFunc(sent, d.Constraints, d.Parameters)
}

type DepRel string

type DepArc struct {
	Modifier int
	Relation DepRel
	Head     int
}

func (arc *DepArc) GetProperty(property string) (string, bool) {
	if property == "l" {
		return arc.Relation, true
	} else {
		return "", false
	}
}
