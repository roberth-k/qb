package qb

import (
	"fmt"

	"github.com/tetratom/qb/internal"
)

type Dialect int

const (
	DialectDefault Dialect = iota
	DialectPq
	DialectGoracle
)

func Lit(s string, args ...interface{}) interface{} {
	return internal.Literal(fmt.Sprintf(s, args...))
}
