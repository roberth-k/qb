package qb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/tetratom/qb/internal"
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
	combiningQuery
	limitExpr
	offsetExpr
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

func (q *Query) writeSQL(s ...string) {
	for i := range s {
		// todo: efficiency...
		s[i] = strings.TrimSpace(s[i])
	}

	q.sql = append(q.sql, s...)
}

func (q *Query) writeArg(v interface{}) {
	q.writeSQL("?")
	q.args = append(q.args, v)
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
	q.writeSQL("DELETE FROM", table)
	return q
}

func (q Query) Values(values ...interface{}) Query {
	return q.ValueTuples(values)
}

func (q Query) ValueTuples(first []interface{}, rest ...[]interface{}) Query {
	q.last = valuesExpr
	q.writeSQL("VALUES")

	all := append([][]interface{}{first}, rest...)

	for _, tuple := range all {
		q.writeSQL("(")
		for i, v := range tuple {
			if i > 0 {
				q.writeSQL(",")
			}

			switch x := v.(type) {
			case internal.Literal:
				q.writeSQL(x.String())
			default:
				q.writeArg(x)
			}
		}

		q.writeSQL(")")
	}

	return q
}

func (q Query) WhereP(pred Predicate) Query {
	if q.last != whereExpr {
		q.writeSQL("WHERE")
	} else {
		q.writeSQL("AND")
	}

	q.last = whereExpr
	q.sql = append(q.sql, pred.w.SQL()...)
	q.args = append(q.args, pred.w.Args()...)
	return q
}

func (q Query) Where(expr string, args ...interface{}) Query {
	return q.WhereP(And(expr, args...))
}

func (q Query) And(expr string, args ...interface{}) Query {
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

func (q Query) appending(t expressionType, expr string, args ...interface{}) Query {
	q.last = t
	q.sql = append(q.sql, expr)
	q.args = append(q.args, args...)
	return q
}

func (q Query) Append(expr string, args ...interface{}) Query {
	return q.appending(q.last, expr, args...)
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
	q.writeSQL(prefix)

	var i, iarg int
	for {
		s := expr[i:]
		if len(s) == 0 {
			break
		}

		j := strings.IndexByte(s, '?')
		if j < 0 {
			q.writeSQL(s)
			break
		}

		q.writeSQL(s[:j])
		switch x := args[iarg].(type) {
		case internal.Literal:
			q.writeSQL(x.String())
		default:
			q.writeArg(x)
		}

		iarg++
		i = j + 1
	}

	return q
}

func (q Query) DefaultValues() Query {
	q.last = defaultValuesExpr
	q.sql = append(q.sql, "DEFAULT", "VALUES")
	return q
}

func (q Query) Union() Query {
	return q.appending(combiningQuery, "UNION")
}

func (q Query) UnionAll() Query {
	return q.appending(combiningQuery, "UNION ALL")
}

func (q Query) Intersect() Query {
	return q.appending(combiningQuery, "INTERSECT")
}

func (q Query) IntersectAll() Query {
	return q.appending(combiningQuery, "INTERSECT ALL")
}

func (q Query) Except() Query {
	return q.appending(combiningQuery, "EXCEPT")
}

func (q Query) ExceptAll() Query {
	return q.appending(combiningQuery, "EXCEPT ALL")
}

func (q Query) Limit(limit int64) Query {
	return q.appending(limitExpr, "LIMIT "+strconv.FormatInt(limit, 10))
}

func (q Query) LimitAll() Query {
	return q.appending(limitExpr, "LIMIT ALL")
}

func (q Query) Offset(offset int64) Query {
	return q.appending(offsetExpr, "OFFSET "+strconv.FormatInt(offset, 10))
}
