package internal_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/tetratom/qb/internal"
)

func TestWriter(t *testing.T) {
	t.Run("WriteExpr", func(t *testing.T) {
		t.Run("IN", func(t *testing.T) {
			var w Writer
			w.WriteExpr("x IN (?, ?)", 1, 2)
			require.Equal(t, []string{"x IN (", "?", ",", "?", ")"}, w.SQL())
			require.Equal(t, []interface{}{1, 2}, w.Args())
		})
	})
}
