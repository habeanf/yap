package Transition

type HasAttributes interface {
	GetProperty(property string) (string, bool)
}

type DepNode struct {
	HeadIndex    int16
	LeftMods     []int16
	RightMods    []int16
	ElementIndex int16
}

func (n *DepNode) SetHead(i int) *DepNode {
	newNode := n.Copy()
	newNode.Head = i
	return newNode
}

func (n *DepNode) AttachLeft(i uint) *DepNode {
	newNode := n.Copy()
	newNode.LeftMods = append(newNode.LeftMods, i)
	return newNode
}

func (n *DepNode) AttachRight(i uint) *DepNode {
	newNode := n.Copy()
	newNode.RightMods = append(newNode.RightMods, i)
}

type DepRel string

type DepArc struct {
	Modifier uint
	Relation DepRel
	Head     uint
}

func (arc *DepArc) GetProperty(property string) (string, bool) {
	if property == "l" {
		return arc.Relation, true
	} else {
		return "", false
	}
}

type Token string

type Sentence []Token

type Stack interface {
	Push(int16)
	Pop() (int16, bool)
	Peek() (int16, bool)
}

type ArcSet interface {
	Push(*DepArc)
	Peek() (*DepArc, bool)
	Get(*DepArc) ([]*DepArc, bool)
}

type Configuration interface {
	HasAttributes

	Init(Sentence)
	Terminal() bool

	Stack() *Stack
	Queue() *Stack
	Arcs() *ArcSet

	Copy() *Configuration
	GetSequence() []Configuration
	SetLastTransition(string)
}

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
