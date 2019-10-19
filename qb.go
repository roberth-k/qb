package qb

import (
	"fmt"
)

type Dialect int

const (
	DialectDefault Dialect = iota
	DialectPq
	DialectGoracle
)

type literal string

func (lit literal) String() string {
	return string(lit)
}

func Lit(s string, args ...interface{}) interface{} {
	return literal(fmt.Sprintf(s, args...))
}
