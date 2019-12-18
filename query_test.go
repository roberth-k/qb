package qb_test

import (
	"fmt"

	"github.com/tetratom/qb"
)

func ExampleQuery_NaturalFullJoin() {
	q := qb.Select("*").From("t1").NaturalFullJoin("t2")
	fmt.Print(q.String())
	// Output: SELECT * FROM t1 NATURAL FULL JOIN t2
}

func ExampleMultiple() {
	q := qb.Multiple(
		qb.Select("*").From("t1"),
		qb.Select("*").From("t2"))
	fmt.Print(q.String())
	// Output: SELECT * FROM t1 ; SELECT * FROM t2 ;
}

func ExampleQuery_GroupBy() {
	q := qb.Select("*").From("t1").GroupBy("a", "b")
	fmt.Print(q.String())
	// Output: SELECT * FROM t1 GROUP BY a, b
}

func ExampleQuery_Having() {
	q := qb.Select("*").From("t1").GroupBy("a", "b").Having(qb.And("a < ?", 100))
	fmt.Print(q.String())
	// Output: SELECT * FROM t1 GROUP BY a, b HAVING a < ?
}

func ExampleQuery_FunctionalMap() {
	var id int64
	joined := true

	q := qb.
		Select("*").
		From("t1").
		Map(func(q qb.Query) qb.Query {
			if joined {
				return q.NaturalJoin("t2")
			}
			return q
		}).
		Where(qb.Predicate{}.
			And(`created_at < now()`).
			Map(func(p qb.Predicate) qb.Predicate {
				if id > 0 {
					return p.And(`id = ?`, id)
				}
				return p
			}))
	fmt.Print(q.String())
	// Output: SELECT * FROM t1 NATURAL JOIN t2 WHERE created_at < now()
}
