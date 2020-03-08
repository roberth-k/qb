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
	combiningQuery
	limitExpr
	offsetExpr
	joinExpr
	groupByExpr
	havingExpr
	usingExpr
)

type Query struct {
	w    sqlWriter
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

func (q Query) FromAs(table, alias string) Query {
	return q.From(As(table, alias))
}

func (q Query) FromSubquery(sq Query) Query {
	q.last = fromExpr
	q.w.WriteSQL("FROM")
	return q.Subquery(sq)
}

func (q Query) Subquery(sq Query) Query {
	q.w.WriteSQL("(")
	q.w.Append(&sq.w)
	q.w.WriteSQL(")")
	return q
}

func Subquery(sq Query) Query {
	return Query{}.Subquery(sq)
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

func DeleteFromAs(table, alias string) Query {
	return DeleteFrom(As(table, alias))
}

func (q Query) DeleteFrom(table string) Query {
	q.last = deleteFromExpr
	q.w.WriteSQL("DELETE FROM", table)
	return q
}

func (q Query) DeleteFromAs(table, alias string) Query {
	return q.DeleteFrom(As(table, alias))
}

func (q Query) Values(values ...interface{}) Query {
	return q.ValueTuples(values)
}

func (q Query) ValueTuples(tuples ...[]interface{}) Query {
	if q.last != valuesExpr {
		q.last = valuesExpr
		q.w.WriteSQL("VALUES")
	} else {
		q.w.WriteSQL(",")
	}

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
			case literal:
				q.w.WriteSQL(x.String())
			case Query:
				q.w.WriteSQL("(")
				q.w.Append(&x.w)
				q.w.WriteSQL(")")
			default:
				q.w.WriteArg(x)
			}
		}

		q.w.WriteSQL(")")
	}

	return q
}

func InsertValuesInto(table string, values Values) Query {
	return Query{}.InsertValuesInto(table, values)
}

func (q Query) InsertValuesInto(table string, values Values) Query {
	columns := make([]string, len(values))
	arguments := make([]interface{}, len(values))
	i := 0
	for k, v := range values {
		columns[i] = k
		arguments[i] = v
		i++
	}

	return q.InsertInto(table, columns...).Values(arguments...)
}

func (q Query) Where(pred Predicate) Query {
	if pred.IsEmpty() {
		return q
	}

	if q.last != whereExpr {
		q.w.WriteSQL("WHERE")
	} else {
		q.w.WriteSQL("AND")
	}

	q.last = whereExpr
	q.w.Append(&pred.w)
	return q
}

func (q Query) Returning(columns ...string) Query {
	q.last = returningExpr
	q.w.WriteSQL("RETURNING")
	for i, column := range columns {
		if i > 0 {
			q.w.WriteSQL(",")
		}

		q.w.WriteSQL(column)
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

func Append(expr string, args ...interface{}) Query {
	return Query{}.Append(expr, args...)
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

func (q Query) SetValues(values Values) Query {
	for k, v := range values {
		q = q.Set(k+` = ?`, v)
	}
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
	q.w.WriteSQL(joinType + " " + table + " ON")
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

func (q Query) JoinAsOn(table, alias string, predicate Predicate) Query {
	return q.joinOn("JOIN", As(table, alias), predicate)
}

func (q Query) JoinUsing(table string, columns ...string) Query {
	return q.joinUsing("JOIN", table, columns...)
}

func (q Query) JoinAsUsing(table, alias string, columns ...string) Query {
	return q.joinUsing("JOIN", As(table, alias), columns...)
}

func (q Query) LeftJoinOn(table string, predicate Predicate) Query {
	return q.joinOn("LEFT JOIN", table, predicate)
}

func (q Query) LeftJoinAsOn(table, alias string, predicate Predicate) Query {
	return q.joinOn("LEFT JOIN", As(table, alias), predicate)
}

func (q Query) LeftJoinUsing(table string, columns ...string) Query {
	return q.joinUsing("LEFT JOIN", table, columns...)
}

func (q Query) LeftJoinAsUsing(table, alias string, columns ...string) Query {
	return q.joinUsing("LEFT JOIN", As(table, alias), columns...)
}

func (q Query) RightJoinOn(table string, predicate Predicate) Query {
	return q.joinOn("RIGHT JOIN", table, predicate)
}

func (q Query) RightJoinAsOn(table, alias string, predicate Predicate) Query {
	return q.joinOn("RIGHT JOIN", As(table, alias), predicate)
}

func (q Query) RightJoinUsing(table string, columns ...string) Query {
	return q.joinUsing("RIGHT JOIN", table, columns...)
}

func (q Query) RightJoinAsUsing(table, alias string, columns ...string) Query {
	return q.joinUsing("RIGHT JOIN", As(table, alias), columns...)
}

func (q Query) FullJoinOn(table string, predicate Predicate) Query {
	return q.joinOn("FULL JOIN", table, predicate)
}

func (q Query) FullJoinAsOn(table, alias string, predicate Predicate) Query {
	return q.joinOn("FULL JOIN", As(table, alias), predicate)
}

func (q Query) FullJoinUsing(table string, columns ...string) Query {
	return q.joinUsing("FULL JOIN", table, columns...)
}

func (q Query) FullJoinAsUsing(table, alias string, columns ...string) Query {
	return q.joinUsing("FULL JOIN", As(table, alias), columns...)
}

func (q Query) NaturalJoin(table string) Query {
	return q.appending(joinExpr, "NATURAL JOIN "+table)
}

func (q Query) NaturalJoinAs(table, alias string) Query {
	return q.appending(joinExpr, "NATURAL JOIN "+As(table, alias))
}

func (q Query) NaturalLeftJoin(table string) Query {
	return q.appending(joinExpr, "NATURAL LEFT JOIN "+table)
}

func (q Query) NaturalLeftJoinAs(table, alias string) Query {
	return q.appending(joinExpr, "NATURAL LEFT JOIN "+As(table, alias))
}

func (q Query) NaturalRightJoin(table string) Query {
	return q.appending(joinExpr, "NATURAL RIGHT JOIN "+table)
}

func (q Query) NaturalRightJoinAs(table, alias string) Query {
	return q.appending(joinExpr, "NATURAL RIGHT JOIN "+As(table, alias))
}

// Appends a NATURAL FULL JOIN clause.
//  ... NATURAL FULL JOIN table
func (q Query) NaturalFullJoin(table string) Query {
	return q.appending(joinExpr, "NATURAL FULL JOIN "+table)
}

func (q Query) NaturalFullJoinAs(table, alias string) Query {
	return q.appending(joinExpr, "NATURAL FULL JOIN "+As(table, alias))
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

func Begin() Query {
	return Query{}.Begin()
}

func (q Query) Begin() Query {
	q.w.WriteSQL("BEGIN")
	return q
}

func Commit() Query {
	return Query{}.Commit()
}

func (q Query) Commit() Query {
	q.w.WriteSQL("COMMIT")
	return q
}

func (q Query) As(alias string) Query {
	q.w.WriteSQL("AS", `"`+alias+`"`)
	return q
}

func (q Query) Map(f func(q Query) Query) Query {
	return f(q)
}

func (q Query) Using(table string) Query {
	q.last = usingExpr
	q.w.WriteSQL("USING", table)
	return q
}

func (q Query) UsingAs(table, alias string) Query {
	return q.Using(As(table, alias))
}
