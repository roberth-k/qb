package qb

type Dialect int

const (
	DialectDefault Dialect = iota
	DialectPq
	DialectGoracle
)
