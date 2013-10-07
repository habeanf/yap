package Util

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
