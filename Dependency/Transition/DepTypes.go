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

type DepNode interface {
	HasAttributes
	Token() string
}

type TaggedDepNode struct {
	Token string
	POS   string
}

func (t *TaggedDepNode) Token() string {
	return t.Token()
}

func (t *TaggedDepNode) GetProperty(prop string) (string, bool) {
	switch prop {
	case "w":
		return t.Token(), true
	case "p":
		return t.POS, true
	default:
		return "", false
	}
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
