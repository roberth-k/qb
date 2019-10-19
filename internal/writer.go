package internal

import (
	"strings"
)

type Writer struct {
	sql  []string
	args []interface{}
}

func (q *Writer) SQL() []string {
	return q.sql
}

func (q *Writer) Args() []interface{} {
	return q.args
}

func (q *Writer) String() string {
	return strings.Join(q.sql, " ")
}

func (q *Writer) Append(sql []string, args []interface{}) {
	q.sql = append(q.sql, sql...)
	q.args = append(q.args, args...)
}

func (q *Writer) WriteSQL(s ...string) {
	q.sql = append(q.sql, s...)
}

func (q *Writer) WriteArg(v interface{}) {
	q.WriteSQL("?")
	q.args = append(q.args, v)
}

func (q *Writer) WriteExpr(expr string, args ...interface{}) {
	var i, iarg int
	for {
		s := expr[i:]
		if len(s) == 0 {
			break
		}

		j := strings.IndexByte(s, '?')
		if j < 0 {
			q.WriteSQL(s)
			break
		}

		q.WriteSQL(strings.TrimSpace(s[:j]))
		switch x := args[iarg].(type) {
		case Literal:
			q.WriteSQL(x.String())
		default:
			q.WriteArg(x)
		}

		iarg++
		i = j + 1
	}
}
