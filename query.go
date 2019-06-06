package qb

import "strings"

type Query struct {
	sql  []string
	args []interface{}
	Dialect
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

func DialectOption(d Dialect) Query {
	return Query{Dialect: d}
}

func (q Query) DialectOption(d Dialect) Query {
	q.Dialect = d
	return q
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

func Select(first string, rest ...string) Query {
	return Query{}.Select(first, rest...)
}

func (q Query) Select(first string, rest ...string) Query {
	q.sql = append(q.sql, "SELECT", first)
	for _, column := range rest {
		q.sql = append(q.sql, ",", column)
	}
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
		for i, column := range columns {
			var start string
			if i == 0 {
				start = "("
			} else {
				start = ","
			}

			q.sql = append(q.sql, start, column)
		}
		q.sql = append(q.sql, ")")
	}

	return q
}

func (q Query) Values(values ...interface{}) Query {
	return q.ValueTuples(values)
}

func (q Query) ValueTuples(first []interface{}, rest ...[]interface{}) Query {
	q.sql = append(q.sql, "VALUES")

	all := append([][]interface{}{first}, rest...)

	for _, tuple := range all {
		// TODO: Yes, this is horribly inefficient.
		// TODO: At minimum, q.sql should be pre-grown.
		q.sql = append(q.sql, "(")
		for i := range tuple {
			if i > 0 {
				q.sql = append(q.sql, ",")
			}
			q.sql = append(q.sql, "?")
		}
		q.sql = append(q.sql, ")")
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

func (q Query) Returning(first string, rest ...string) Query {
	q.sql = append(q.sql, "RETURNING", first)
	for _, column := range rest {
		q.sql = append(q.sql, ",", column)
	}
	return q
}
