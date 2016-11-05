package util

import (
	"encoding/gob"
	"fmt"
	"log"
	"sync"
)

func init() {
	gob.Register([2]string{})
}

type EnumSet struct {
	mu     sync.RWMutex
	Enum   map[interface{}]int
	Index  []interface{}
	Frozen bool
}

func (e *EnumSet) RebuildIndex() {
	e.mu.Lock()
	defer e.mu.Unlock()
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
	e.mu.RLock()
	defer e.mu.RUnlock()
	enum, exists := e.Enum[value]
	return enum, exists
}

func (e *EnumSet) ValueOf(index int) interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if index < 0 {
		panic("Negative index requested")
	}
	if len(e.Index) != len(e.Enum) {
		log.Println("Rebuilding index!")
		e.RebuildIndex()
	}
	if len(e.Index) <= index {
		panic("Unknown index requested: " + fmt.Sprintf("%v of %v", index, len(e.Index)))
	}
	return e.Index[index]
}

func (e *EnumSet) Len() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.Index)
}

func (e *EnumSet) Print() {
	for i, v := range e.Index {
		fmt.Printf("%v: %v\n", i, v)
	}
}

func NewEnumSet(capacity int) *EnumSet {
	e := &EnumSet{
		sync.RWMutex{},
		make(map[interface{}]int, capacity),
		make([]interface{}, 0, capacity),
		false,
	}
	return e
}
