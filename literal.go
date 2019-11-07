package qb

type literal string

func (lit literal) String() string {
	return string(lit)
}
