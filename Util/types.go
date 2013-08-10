package Util

type Equaler interface {
	Equal(Equaler) bool
}
