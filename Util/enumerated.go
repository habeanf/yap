package Util

import "sync"

type EnumSet struct {
	mu     sync.Mutex
	Enum   map[interface{}]int
	Index  []interface{}
	Frozen bool
}

func (e *EnumSet) MapAdd(value interface{}) (int, bool) {
	if e.Frozen {
		panic("Cannot add value to frozen enum set")
	}
	enum, exists := e.Enum[value]
	if exists {
		return enum, false
	}
	enum = len(e.Enum)
	e.Enum[value] = enum
	return enum, true
}

func (e *EnumSet) RebuildIndex() {
	e.Index = make([]interface{}, len(e.Enum))
	for k, v := range e.Enum {
		e.Index[v] = k
	}
}

func (e *EnumSet) Add(value interface{}) (int, bool) {
	if e.Frozen {
		panic("Cannot add value to frozen enum set")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	enum, exists := e.Enum[value]
	if exists {
		return enum, false
	}
	enum = len(e.Index)
	e.Enum[value] = enum
	e.Index = append(e.Index, value)
	return enum, true
}

func (e *EnumSet) IndexOf(value interface{}) (int, bool) {
	enum, exists := e.Enum[value]
	return enum, exists
}

func (e *EnumSet) ValueOf(index int) interface{} {
	if index < 0 {
		panic("Negative index requested")
	}
	if len(e.Index) != len(e.Enum) {
		e.RebuildIndex()
	}
	if len(e.Index) <= index {
		panic("Unknown index requested")
	}
	return e.Index[index]
}

func (e *EnumSet) Len() int {
	return len(e.Index)
}

func NewEnumSet(capacity int) *EnumSet {
	e := &EnumSet{
		sync.Mutex{},
		make(map[interface{}]int, capacity),
		make([]interface{}, 0, capacity),
		false,
	}
	return e
}
