package qb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type expressionType int

const (
	anyExpr expressionType = iota
	defaultValuesExpr
	fromExpr
	insertIntoExpr
	orderByExpr
	returningExpr
	selectExpr
	setExpr
	updateExpr
	valuesExpr
	whereExpr
	withExpr
	deleteFromExpr
)

type Query struct {
	sql  []string
	args []interface{}
	last expressionType
	str  string
	Dialect
}

func (q Query) String() string {
	return q.SQL()
}

// TODO: Support escapes.
var placeholderPattern = regexp.MustCompile(`\?`)

func (q *Query) SQL() string {
	if q.str != "" {
		return q.str
	}

	sql := strings.Join(q.sql, " ")

	var prefix string
	switch q.Dialect {
	case DialectDefault:
		return sql
	case DialectPq:
		prefix = "$"
	case DialectGoracle:
		prefix = ":"
	default:
		panic(fmt.Errorf("unrecognised dialect %d", q.Dialect))
	}

	i := 0
	q.str = placeholderPattern.ReplaceAllStringFunc(
		sql, func(match string) string {
			i += 1
			return prefix + strconv.Itoa(i)
		})

	return q.str
}

func (q *Query) Args() []interface{} {
	args := make([]interface{}, len(q.args))
	copy(args, q.args)
	return args
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
	prefix := "WITH"
	if q.last == withExpr {
		prefix = ","
	}

	q.last = withExpr
	q.sql = append(q.sql, prefix, name, "AS", "(")
	q.sql = append(q.sql, query.sql...)
	q.sql = append(q.sql, ")")
	q.args = append(q.args, query.args...)
	return q
}

func Select(first string, rest ...string) Query {
	return Query{}.Select(first, rest...)
}

func (q Query) Select(first string, rest ...string) Query {
	q.last = selectExpr
	q.sql = append(q.sql, "SELECT", first)
	for _, column := range rest {
		q.sql = append(q.sql, ",", column)
	}
	return q
}

func (q Query) From(expr string) Query {
	q.last = fromExpr
	q.sql = append(q.sql, "FROM", expr)
	return q
}

func InsertInto(expr string, columns ...string) Query {
	return Query{}.InsertInto(expr, columns...)
}

func (q Query) InsertInto(expr string, columns ...string) Query {
	q.last = insertIntoExpr
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

func DeleteFrom(table string) Query {
	return Query{}.DeleteFrom(table)
}

func (q Query) DeleteFrom(table string) Query {
	q.last = deleteFromExpr
	q.sql = append(q.sql, "DELETE FROM", table)
	return q
}

func (q Query) Values(values ...interface{}) Query {
	return q.ValueTuples(values)
}

func (q Query) ValueTuples(first []interface{}, rest ...[]interface{}) Query {
	q.last = valuesExpr
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

func (q Query) WhereP(pred Predicate) Query {
	q.last = whereExpr
	q.sql = append(q.sql, "WHERE")
	q.sql = append(q.sql, pred.sql...)
	q.args = append(q.args, pred.args...)
	return q
}

func (q Query) Where(expr string, args ...interface{}) Query {
	return q.WhereP(And(expr, args...))
}

func (q Query) Returning(first string, rest ...string) Query {
	q.last = returningExpr
	q.sql = append(q.sql, "RETURNING", first)
	for _, column := range rest {
		q.sql = append(q.sql, ",", column)
	}
	return q
}

func (q Query) OrderBy(first string, rest ...string) Query {
	q.last = orderByExpr
	q.sql = append(q.sql, "ORDER BY", first)
	for _, column := range rest {
		q.sql = append(q.sql, ",", column)
	}
	return q
}

func (q Query) Append(expr string, args ...interface{}) Query {
	q.sql = append(q.sql, expr)
	q.args = append(q.args, args...)
	return q
}

func Update(table string) Query {
	return Query{}.Update(table)
}

func (q Query) Update(table string) Query {
	q.last = updateExpr
	q.sql = append(q.sql, "UPDATE", table)
	return q
}

func (q Query) Set(expr string, args ...interface{}) Query {
	prefix := "SET"
	if q.last == setExpr {
		prefix = ","
	}

	q.last = setExpr
	q.sql = append(q.sql, prefix, expr)
	q.args = append(q.args, args...)
	return q
}

func (q Query) DefaultValues() Query {
	q.last = defaultValuesExpr
	q.sql = append(q.sql, "DEFAULT", "VALUES")
	return q
}
