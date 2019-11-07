package qb_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratom/qb"
)

func TestQuery(t *testing.T) {
	tests := []struct {
		name  string
		expr  string
		args  []interface{}
		query func() qb.Query
	}{
		{
			name: "simple select",
			expr: `SELECT * FROM my_table`,
			args: []interface{}{},
			query: func() qb.Query {
				return qb.
					Select("*").
					From("my_table")
			},
		},
		{
			name: "simple select with compound predicate",
			expr: `SELECT * FROM my_table WHERE x = 1 AND y = ?`,
			args: []interface{}{2},
			query: func() qb.Query {
				return qb.
					Select("*").
					From("my_table").
					Where(qb.
						And("x = 1").
						And("y = ?", 2))
			},
		},
		{
			name: "simple select with compound predicate and pq dialect",
			expr: `SELECT * FROM my_table WHERE x = $1 AND y = $2`,
			args: []interface{}{1, 2},
			query: func() qb.Query {
				return qb.
					DialectOption(qb.DialectPq).
					Select("*").
					From("my_table").
					Where(qb.
						And("x = ?", 1).
						And("y = ?", 2))
			},
		},
		{
			name: "simple select with nested predicate",
			expr: `SELECT a , b FROM my_table WHERE x = ? AND ( y = ? OR z = ? )`,
			args: []interface{}{1, 2, 3},
			query: func() qb.Query {
				return qb.
					Select("a", "b").
					From("my_table").
					Where(qb.
						And("x = ?", 1).
						AndP(qb.
							And("y = ?", 2).
							Or("z = ?", 3)))
			},
		},
		{
			name: "simple insert with arguments",
			expr: `INSERT INTO my_table ( a , b , c ) VALUES ( ? , ? , ? )`,
			args: []interface{}{"a", "b", "c"},
			query: func() qb.Query {
				return qb.
					InsertInto("my_table", "a", "b", "c").
					Values("a", "b", "c")
			},
		},
		{
			name: "simple insert with arguments and tuples",
			expr: `INSERT INTO my_table ( a ) VALUES ( ? ) , ( ? )`,
			args: []interface{}{1, 2},
			query: func() qb.Query {
				return qb.
					InsertInto("my_table", "a").
					ValueTuples([]interface{}{1}, []interface{}{2})
			},
		},
		{
			name: "with statement",
			expr: `WITH stmt1 AS ( INSERT INTO my_table ( a , b , c ) VALUES ( ? , ? , ? ) ) SELECT a AS "foo.bar" FROM my_table WHERE a = ?`,
			args: []interface{}{1, 2, 3, 1},
			query: func() qb.Query {
				return qb.
					With("stmt1", qb.
						InsertInto("my_table", "a", "b", "c").
						Values(1, 2, 3)).
					Select(`a AS "foo.bar"`).
					From("my_table").
					Where(qb.
						And("a = ?", 1))
			},
		},
		{
			name: "multiple with statements",
			expr: `WITH stmt1 AS ( INSERT INTO my_table DEFAULT VALUES ) , stmt2 AS ( INSERT INTO other_table DEFAULT VALUES ) SELECT * FROM my_table`,
			args: []interface{}{},
			query: func() qb.Query {
				return qb.
					With("stmt1", qb.
						InsertInto("my_table").
						DefaultValues()).
					With("stmt2", qb.
						InsertInto("other_table").
						DefaultValues()).
					Select("*").
					From("my_table")
			},
		},
		{
			name: "returning",
			expr: `INSERT INTO my_table ( a ) VALUES ( ? ) RETURNING a`,
			args: []interface{}{1},
			query: func() qb.Query {
				return qb.
					InsertInto("my_table", "a").
					Values(1).
					Returning("a")
			},
		},
		{
			name: "order by",
			expr: `SELECT * FROM my_table ORDER BY a ASC , b DESC`,
			args: []interface{}{},
			query: func() qb.Query {
				return qb.
					Select("*").
					From("my_table").
					OrderBy("a ASC", "b DESC")
			},
		},
		{
			name: "update table and where",
			expr: `UPDATE my_table SET foo = 1 , bar = ? WHERE a = ?`,
			args: []interface{}{"a", 2},
			query: func() qb.Query {
				return qb.
					Update("my_table").
					Set("foo = 1").Set("bar = ?", "a").
					Where(qb.And("a = ?", 2))
			},
		},
		{
			name: "limit and offset",
			expr: `SELECT * FROM my_table ORDER BY a LIMIT 10 OFFSET 5`,
			args: []interface{}{},
			query: func() qb.Query {
				return qb.
					Select("*").
					From("my_table").
					OrderBy("a").
					Limit(10).
					Offset(5)
			},
		},
		{
			name: "insert with literal",
			expr: `INSERT INTO my_table ( a , b , c ) VALUES ( ? , ? , now() )`,
			args: []interface{}{1, 2},
			query: func() qb.Query {
				return qb.
					InsertInto("my_table", "a", "b", "c").
					Values(1, 2, qb.Lit("now()"))
			},
		},
		{
			name: "update with literal",
			expr: `UPDATE my_table SET updated = now() WHERE id = ?`,
			args: []interface{}{1},
			query: func() qb.Query {
				return qb.
					Update("my_table").
					Set("updated = ?", qb.Lit("now()")).
					Where(qb.And("id = ?", 1))
			},
		},
		{
			name: "query with literal",
			expr: `SELECT * FROM my_table WHERE a = ? AND b = now()`,
			args: []interface{}{1},
			query: func() qb.Query {
				return qb.
					Select("*").
					From("my_table").
					Where(qb.And("a = ?", 1).And("b = ?", qb.Lit("now()")))
			},
		},
		{
			name: "readme example",
			expr: `SELECT * FROM my_table WHERE id = ? OR time < now() ORDER BY time ASC LIMIT 10`,
			args: []interface{}{1},
			query: func() qb.Query {
				return qb.
					Select("*").From("my_table").
					Where(qb.
						And("id = ?", 1).
						Or("time < ?", qb.Lit("now()"))).
					OrderBy("time ASC").
					Limit(10)
			},
		},
		{
			name: "IN predicate",
			expr: `SELECT * FROM my_table WHERE state IN ( ? , ? ) AND created_time < ? ORDER BY created_time DESC LIMIT 1 OFFSET 5`,
			args: []interface{}{1, 2, 3},
			query: func() qb.Query {
				return qb.
					Select("*").From("my_table").
					Where(qb.
						And("state IN (?, ?)", 1, 2).
						And("created_time < ?", 3)).
					OrderBy("created_time DESC").
					Limit(1).
					Offset(5)
			},
		},
		{
			name: "JOIN ON",
			expr: `SELECT * FROM t1 JOIN t2 ON t1.id = t2.id`,
			args: []interface{}{},
			query: func() qb.Query {
				return qb.Select("*").From("t1").JoinOn("t2", qb.And("t1.id = t2.id"))
			},
		},
		{
			name: "JOIN USING",
			expr: `SELECT * FROM t1 JOIN t2 USING ( a , b )`,
			args: []interface{}{},
			query: func() qb.Query {
				return qb.Select("*").From("t1").JoinUsing("t2", "a", "b")
			},
		},
		{
			name: "Multiple queries",
			expr: "SELECT * FROM t1 WHERE a = $1 ; SELECT * FROM t2 WHERE b = $2 ;",
			args: []interface{}{1, 2},
			query: func() qb.Query {
				return qb.Multiple(
					qb.Select("*").From("t1").Where(qb.And("a = ?", 1)),
					qb.Select("*").From("t2").Where(qb.And("b = ?", 2))).
					DialectOption(qb.DialectPq)
			},
		},
		{
			name: "GROUP BY ... HAVING ...",
			expr: "SELECT * FROM t1 GROUP BY a, b HAVING a < 500 AND b > ?",
			args: []interface{}{1},
			query: func() qb.Query {
				return qb.
					Select("*").From("t1").
					GroupBy("a", "b").
					Having(qb.And("a < 500").And("b > ?", 1))
			},
		},
		{
			name: "SelectColumn",
			expr: `INSERT INTO t1 ( a , b , c , d ) SELECT a ,  ? , c ,  ? FROM t2 WHERE x = ?`,
			args: []interface{}{1, 2, 3},
			query: func() qb.Query {
				return qb.
					InsertInto("t1", "a", "b", "c", "d").
					Select("a").
					SelectColumn("?", 1).
					Select("c").
					SelectColumn("?", 2).
					From("t2").
					Where(qb.And("x = ?", 3))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := test.query()
			require.Equal(t, test.expr, q.SQL())
			require.Equal(t, test.args, q.Args())
		})
	}
}
