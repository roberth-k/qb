package qb

import "strings"

type Predicate struct {
	sql   []string
	args  []interface{}
	count int
}

func (p Predicate) String() string {
	return strings.Join(p.sql, " ")
}

func Pred(expr string, args ...interface{}) Predicate {
	return Predicate{}.And(expr, args...)
}

func And(expr string, args ...interface{}) Predicate {
	return Pred(expr, args...)
}

func (my Predicate) And(expr string, args ...interface{}) Predicate {
	if my.count > 0 {
		my.sql = append(my.sql, "AND")
	}

	my.count += 1
	my.sql = append(my.sql, expr)
	my.args = append(my.args, args...)
	return my
}

func AndP(predicate Predicate) Predicate {
	return Predicate{}.AndP(predicate)
}

func (my Predicate) AndP(predicate Predicate) Predicate {
	if my.count > 0 {
		my.sql = append(my.sql, "AND")
	}

	my.count += 1
	my.sql = append(my.sql, "(")
	my.sql = append(my.sql, predicate.sql...)
	my.sql = append(my.sql, ")")
	my.args = append(my.args, predicate.args...)
	return my
}

func (my Predicate) Or(expr string, args ...interface{}) Predicate {
	if my.count > 0 {
		my.sql = append(my.sql, "OR")
	}

	my.count += 1
	my.sql = append(my.sql, expr)
	my.args = append(my.args, args...)
	return my
}

func (my Predicate) OrP(predicate Predicate) Predicate {
	if my.count > 0 {
		my.sql = append(my.sql, "OR")
	}

	my.count += 1
	my.sql = append(my.sql, "(")
	my.sql = append(my.sql, predicate.sql...)
	my.sql = append(my.sql, ")")
	my.args = append(my.args, predicate.args...)
	return my
}
