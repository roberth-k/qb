package internal

type Literal string

func (lit Literal) String() string {
	return string(lit)
}
