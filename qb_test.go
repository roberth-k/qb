package qb_test

import (
	"github.com/stretchr/testify/require"
	"github.com/tetratom/qb"
	"testing"
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
			args: nil,
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
						AndS("x = 1").
						AndS("y = ?", 2))
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
						AndS("x = ?", 1).
						AndS("y = ?", 2))
			},
		},
		{
			name: "simple select with nested predicate",
			expr: `SELECT * FROM my_table WHERE x = ? AND ( y = ? OR z = ? )`,
			args: []interface{}{1, 2, 3},
			query: func() qb.Query {
				return qb.
					Select("*").
					From("my_table").
					Where(qb.
						AndS("x = ?", 1).
						AndP(qb.
							AndS("y = ?", 2).
							OrS("z = ?", 3)))
			},
		},
		{
			name: "simple insert with arguments",
			expr: `INSERT INTO my_table ( a, b, c ) VALUES ( ?, ?, ? )`,
			args: []interface{}{"a", "b", "c"},
			query: func() qb.Query {
				return qb.
					InsertInto("my_table", "a", "b", "c").
					Values("a", "b", "c")
			},
		},
		{
			name: "with statement",
			expr: `WITH stmt1 AS ( INSERT INTO my_table ( a, b, c ) VALUES ( ?, ?, ? ) ) SELECT a AS "foo.bar" FROM my_table WHERE a = ?`,
			args: []interface{}{1, 2, 3, 1},
			query: func() qb.Query {
				return qb.
					With("stmt1", qb.
						InsertInto("my_table", "a", "b", "c").
						Values(1, 2, 3)).
					Select(`a AS "foo.bar"`).
					From("my_table").
					Where(qb.
						AndS("a = ?", 1))
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
