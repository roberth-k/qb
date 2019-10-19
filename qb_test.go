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
					WhereP(qb.
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
					WhereP(qb.
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
					WhereP(qb.
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
					WhereP(qb.
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
					Where("a = ?", 2)
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
					Where("id = ?", 1)
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
					WhereP(qb.And("a = ?", 1).And("b = ?", qb.Lit("now()")))
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
