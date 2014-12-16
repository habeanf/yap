package util

import "io"

type Equaler interface {
	Equal(Equaler) bool
}

type Persist interface {
	Read(reader io.Reader)
	Write(writer io.Writer)
}

type Format interface {
	Format(value interface{}) string
}

type Generic struct {
	Key   string
	Value interface{}
}

type ByGeneric []Generic

func (b ByGeneric) Len() int           { return len(b) }
func (b ByGeneric) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByGeneric) Less(i, j int) bool { return b[i].Key < b[j].Key }
