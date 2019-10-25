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
