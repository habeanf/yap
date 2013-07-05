package Dependency

import "strconv"

type Parser interface {
	Init()
}

type HasProperties interface {
	func GetProperty(property string) (string, bool)
}

type DepNode struct {
	HeadIndex       uint
	ModifierIndices []uint
	LeftMods        []uint
	RightMods       []uint
	ElementIndex    uint
}

type DepRel string

type DepArc struct {
	Modifier uint
	Head     uint
	Relation DepRel
}

func (arc *DepArc) GetProperty(property string) (string, bool) {
	if property == "l" {
		return arc.Relation, true
	} else {
		return "", false
	}
}

type Configuration struct {
	Stack    []DepNode
	Queue    []DepNode
	Arcs     []DepArc
	Elements *[]HasProperties
}

func (c *Configuration) GetProperty(property string) (string,bool) {
	if property == "d" {
		return "1", true
	} else {
		return "", false 
	}
}
func (c *Configuration) GetSource(source string) *interface{} {
	switch source {
	case "N":
		return &(c.Queue)
	case "S":
		return &(c.Stack)
	}
	return nil
}
func (c *Configuration) GetLocation(currentTarget interface{}, location string) (*HasProperties,bool) {
	switch t:= currentTarget.(type) {
	default:
		return nil, false
	case *[]DepNode:

		offset, err := strconv.ParseInt(currentLocation, 10, 0)
		if !err {
			
		}
	case *DepNode:
		switch location[0] {
		case "h":
			return 
		case "l":
		case "r":
		}
	case *DepArc:
		return currentTarget, true
	}	
	if len(location) == 1 {
		return currentTarget, true
	}
}