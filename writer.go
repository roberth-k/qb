package qb

import (
	"strings"
)

type sqlWriter struct {
	sql  []string
	args []interface{}
}

func (q *sqlWriter) SQL() []string {
	return q.sql
}

func (q *sqlWriter) Args() []interface{} {
	return q.args
}

func (q *sqlWriter) String() string {
	return strings.Join(q.sql, " ")
}

func (q *sqlWriter) Append(w *sqlWriter) {
	q.sql = append(q.sql, w.sql...)
	q.args = append(q.args, w.args...)
}

func (q *sqlWriter) WriteSQL(s ...string) {
	q.sql = append(q.sql, s...)
}

func (q *sqlWriter) WriteArg(v interface{}) {
	q.WriteSQL("?")
	q.args = append(q.args, v)
}

func (q *sqlWriter) WriteExpr(expr string, args ...interface{}) {
	var i, iarg int
	for {
		s := expr[i:]
		if len(s) == 0 {
			break
		}

		j := strings.IndexRune(s, '?')
		if j < 0 {
			q.WriteSQL(s)
			break
		}

		q.WriteSQL(strings.TrimSpace(s[:j]))
		switch x := args[iarg].(type) {
		case literal:
			q.WriteSQL(x.String())
		default:
			q.WriteArg(x)
		}

		iarg++
		i += j + 1
	}
}
