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
	sql1 := make([]string, 0, len(q.sql)+len(w.sql))
	sql1 = append(sql1, q.sql...)
	sql1 = append(sql1, w.sql...)

	args1 := make([]interface{}, 0, len(q.args)+len(w.args))
	args1 = append(args1, q.args...)
	args1 = append(args1, w.args...)

	q.sql, q.args = sql1, args1
}

func (q *sqlWriter) WriteSQL(s ...string) {
	sql1 := make([]string, 0, len(q.sql)+len(s))
	sql1 = append(sql1, q.sql...)
	sql1 = append(sql1, s...)
	q.sql = sql1
}

func (q *sqlWriter) WriteArg(v interface{}) {
	q.WriteSQL("?")

	args1 := make([]interface{}, 0, len(q.args)+1)
	args1 = append(args1, q.args...)
	args1 = append(args1, v)
	q.args = args1
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
		case Query:
			q.WriteSQL("(")
			q.Append(&x.w)
			q.WriteSQL(")")
		default:
			q.WriteArg(x)
		}

		iarg++
		i += j + 1
	}
}
