package qb

type Predicate struct {
	w     sqlWriter
	count int
}

func (p Predicate) String() string {
	return p.w.String()
}

func (p Predicate) IsEmpty() bool {
	return p.count == 0
}

func Pred(expr string, args ...interface{}) Predicate {
	return Predicate{}.And(expr, args...)
}

func And(expr string, args ...interface{}) Predicate {
	return Pred(expr, args...)
}

func (my Predicate) And(expr string, args ...interface{}) Predicate {
	if my.count > 0 {
		my.w.WriteSQL("AND")
	}

	my.count += 1
	my.w.WriteExpr(expr, args...)
	return my
}

func AndP(predicate Predicate) Predicate {
	return Predicate{}.AndP(predicate)
}

func (my Predicate) AndP(predicate Predicate) Predicate {
	if predicate.IsEmpty() {
		return my
	}

	if my.count > 0 {
		my.w.WriteSQL("AND")
	}

	my.count += 1
	my.w.WriteSQL("(")
	my.w.Append(&predicate.w)
	my.w.WriteSQL(")")
	return my
}

func (my Predicate) Or(expr string, args ...interface{}) Predicate {
	if my.count > 0 {
		my.w.WriteSQL("OR")
	}

	my.count += 1
	my.w.WriteExpr(expr, args...)
	return my
}

func (my Predicate) OrP(predicate Predicate) Predicate {
	if predicate.IsEmpty() {
		return my
	}

	if my.count > 0 {
		my.w.WriteSQL("OR")
	}

	my.count += 1
	my.w.WriteSQL("(")
	my.w.Append(&predicate.w)
	my.w.WriteSQL(")")
	return my
}
