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
	joinExpr
	groupByExpr
	havingExpr
)

type Query struct {
	w    internal.Writer
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

	sql := q.w.String()

	var prefix string
	switch q.Dialect {
	case DialectDefault:
		return sql
	case DialectPq:
		prefix = "$"
	case DialectGoracle:
		prefix = ":"
	case DialectMssql:
		prefix = "@p"
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
	in := q.w.Args()
	args := make([]interface{}, len(in))
	copy(args, in)
	return args
}

func (q Query) Build() (string, []interface{}) {
	return q.SQL(), q.Args()
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
	q.w.WriteSQL(prefix, name, "AS", "(")
	q.w.Append(&query.w)
	q.w.WriteSQL(")")
	return q
}

func Select(columns ...string) Query {
	return Query{}.Select(columns...)
}

func (q Query) Select(columns ...string) Query {
	if q.last != selectExpr {
		q.last = selectExpr
		q.w.WriteSQL("SELECT")
	} else {
		q.w.WriteSQL(",")
	}

	for i, column := range columns {
		if i > 0 {
			q.w.WriteSQL(",")
		}

		q.w.WriteSQL(column)
	}
	return q
}

func SelectColumn(expr string, args ...interface{}) Query {
	return Query{}.SelectColumn(expr, args...)
}

func (q Query) SelectColumn(expr string, args ...interface{}) Query {
	if q.last != selectExpr {
		q.last = selectExpr
		q.w.WriteSQL("SELECT")
	} else {
		q.w.WriteSQL(",")
	}

	q.w.WriteExpr(expr, args...)
	return q
}

func (q Query) From(expr string) Query {
	q.last = fromExpr
	q.w.WriteSQL("FROM", expr)
	return q
}

func InsertInto(expr string, columns ...string) Query {
	return Query{}.InsertInto(expr, columns...)
}

func (q Query) InsertInto(expr string, columns ...string) Query {
	q.last = insertIntoExpr
	q.w.WriteSQL("INSERT INTO", expr)

	if len(columns) > 0 {
		for i, column := range columns {
			var start string
			if i == 0 {
				start = "("
			} else {
				start = ","
			}

			q.w.WriteSQL(start, column)
		}
		q.w.WriteSQL(")")
	}

	return q
}

func DeleteFrom(table string) Query {
	return Query{}.DeleteFrom(table)
}

func (q Query) DeleteFrom(table string) Query {
	q.last = deleteFromExpr
	q.w.WriteSQL("DELETE FROM", table)
	return q
}

func (q Query) Values(values ...interface{}) Query {
	return q.ValueTuples(values)
}

func (q Query) ValueTuples(tuples ...[]interface{}) Query {
	q.last = valuesExpr
	q.w.WriteSQL("VALUES")

	for i, tuple := range tuples {
		if i > 0 {
			q.w.WriteSQL(",")
		}

		q.w.WriteSQL("(")

		for i, v := range tuple {
			if i > 0 {
				q.w.WriteSQL(",")
			}

			switch x := v.(type) {
			case internal.Literal:
				q.w.WriteSQL(x.String())
			default:
				q.w.WriteArg(x)
			}
		}

		q.w.WriteSQL(")")
	}

	return q
}

func (q Query) Where(pred Predicate) Query {
	if q.last != whereExpr {
		q.w.WriteSQL("WHERE")
	} else {
		q.w.WriteSQL("AND")
	}

	q.last = whereExpr
	q.w.Append(&pred.w)
	return q
}

func (q Query) Returning(first string, rest ...string) Query {
	q.last = returningExpr
	q.w.WriteSQL("RETURNING", first)
	for _, column := range rest {
		q.w.WriteSQL(",", column)
	}
	return q
}

func (q Query) OrderBy(first string, rest ...string) Query {
	q.last = orderByExpr
	q.w.WriteSQL("ORDER BY", first)
	for _, column := range rest {
		q.w.WriteSQL(",", column)
	}
	return q
}

func (q Query) appending(t expressionType, expr string, args ...interface{}) Query {
	q.last = t
	q.w.WriteExpr(expr, args...)
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
	q.w.WriteSQL("UPDATE", table)
	return q
}

func (q Query) Set(expr string, args ...interface{}) Query {
	prefix := "SET"
	if q.last == setExpr {
		prefix = ","
	}

	q.last = setExpr
	q.w.WriteSQL(prefix)
	q.w.WriteExpr(expr, args...)
	return q
}

func (q Query) DefaultValues() Query {
	q.last = defaultValuesExpr
	q.w.WriteSQL("DEFAULT VALUES")
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

func (q Query) joinOn(joinType string, table string, predicate Predicate) Query {
	q.last = joinExpr
	q.w.WriteSQL(joinType+" "+table+" ON")
	q.w.Append(&predicate.w)
	return q
}

func (q Query) joinUsing(joinType string, table string, columns ...string) Query {
	q.last = joinExpr
	q.w.WriteSQL(joinType + " " + table + " USING (")
	for i, column := range columns {
		if i > 0 {
			q.w.WriteSQL(",")
		}
		q.w.WriteSQL(column)
	}
	q.w.WriteSQL(")")
	return q
}

func (q Query) JoinOn(table string, predicate Predicate) Query {
	return q.joinOn("JOIN", table, predicate)
}

func (q Query) JoinUsing(table string, columns ...string) Query {
	return q.joinUsing("JOIN", table, columns...)
}

func (q Query) LeftJoinOn(table string, predicate Predicate) Query {
	return q.joinOn("LEFT JOIN", table, predicate)
}

func (q Query) LeftJoinUsing(table string, columns ...string) Query {
	return q.joinUsing("LEFT JOIN", table, columns...)
}

func (q Query) RightJoinOn(table string, predicate Predicate) Query {
	return q.joinOn("RIGHT JOIN", table, predicate)
}

func (q Query) RightJoinUsing(table string, columns ...string) Query {
	return q.joinUsing("RIGHT JOIN", table, columns...)
}

func (q Query) FullJoinOn(table string, predicate Predicate) Query {
	return q.joinOn("FULL JOIN", table, predicate)
}

func (q Query) FullJoinUsing(table string, columns ...string) Query {
	return q.joinUsing("FULL JOIN", table, columns...)
}

func (q Query) NaturalJoin(table string) Query {
	return q.appending(joinExpr, "NATURAL JOIN "+table)
}

func (q Query) NaturalLeftJoin(table string) Query {
	return q.appending(joinExpr, "NATURAL LEFT JOIN "+table)
}

func (q Query) NaturalRightJoin(table string) Query {
	return q.appending(joinExpr, "NATURAL RIGHT JOIN "+table)
}

// Appends a NATURAL FULL JOIN clause.
//  ... NATURAL FULL JOIN table
func (q Query) NaturalFullJoin(table string) Query {
	return q.appending(joinExpr, "NATURAL FULL JOIN "+table)
}

// Creates a query with multiple statements.
//  qs0; [qs1; [qs2; ...]]
func Multiple(qs ...Query) Query {
	var out Query
	for _, in := range qs {
		out.w.Append(&in.w)
		out.w.WriteSQL(";")
	}
	return out
}

// Appends a GROUP BY clause.
//  ... GROUP BY field0[, field1[, ...]].
func (q Query) GroupBy(fields ...string) Query {
	return q.appending(groupByExpr, "GROUP BY "+strings.Join(fields, ", "))
}

// Appends a HAVING clause. This should follow a GROUP BY clause.
//  ... HAVING predicate
func (q Query) Having(predicate Predicate) Query {
	q.last = havingExpr
	q.w.WriteSQL("HAVING")
	q.w.Append(&predicate.w)
	return q
}

func (q Query) ForUpdate() Query {
	q.w.WriteSQL("FOR UPDATE")
	return q
}

func (q Query) ForShare() Query {
	q.w.WriteSQL("FOR SHARE")
	return q
}
