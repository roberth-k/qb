package qb

import "strings"

type Query struct {
	sql  []string
	args []interface{}
}

func (q Query) String() string {
	return q.SQL()
}

func (q Query) SQL() string {
	return strings.Join(q.sql, " ")
}

func (q Query) Args() []interface{} {
	return q.args
}

func With(name string, query Query) Query {
	return Query{}.With(name, query)
}

func (q Query) With(name string, query Query) Query {
	q.sql = append(q.sql, "WITH", name, "AS", "(")
	q.sql = append(q.sql, query.sql...)
	q.sql = append(q.sql, ")")
	q.args = append(q.args, query.args...)
	return q
}

func Select(expr string) Query {
	return Query{}.Select(expr)
}

func (q Query) Select(expr string) Query {
	q.sql = append(q.sql, "SELECT", expr)
	return q
}

func (q Query) From(expr string) Query {
	q.sql = append(q.sql, "FROM", expr)
	return q
}

func InsertInto(expr string, columns ...string) Query {
	return Query{}.InsertInto(expr, columns...)
}

func (q Query) InsertInto(expr string, columns ...string) Query {
	q.sql = append(q.sql, "INSERT INTO", expr)

	if len(columns) > 0 {
		q.sql = append(q.sql, "(", strings.Join(columns, ", "), ")")
	}

	return q
}

func placeholders(n int) string {
	qm, sep := "?", ", "
	b := strings.Builder{}
	b.Grow(n*len(qm) + (n-1)*len(sep))
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(qm)
	}
	return b.String()
}

func (q Query) Values(values ...interface{}) Query {
	return q.ValueTuples(values)
}

func (q Query) ValueTuples(first []interface{}, rest ...[]interface{}) Query {
	q.sql = append(q.sql, "VALUES")
	q.sql = append(q.sql, "(", placeholders(len(first)), ")")
	q.args = append(q.args, first...)

	for _, tuple := range rest {
		q.sql = append(q.sql, "(", placeholders(len(tuple)), ")")
		q.args = append(q.args, tuple...)
	}

	return q
}

func (q Query) Where(pred Predicate) Query {
	q.sql = append(q.sql, "WHERE")
	q.sql = append(q.sql, pred.sql...)
	q.args = append(q.args, pred.args...)
	return q
}
