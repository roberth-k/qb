qb
===

[![GoDoc](https://godoc.org/github.com/tetratom/qb?status.svg)](https://godoc.org/github.com/tetratom/qb)
[![CircleCI](https://circleci.com/gh/tetratom/qb.svg?style=svg)](https://circleci.com/gh/tetratom/qb)
[![Go Report Card](https://goreportcard.com/badge/github.com/tetratom/qb)](https://goreportcard.com/report/github.com/tetratom/qb)

qb is a simple SQL query builder for Golang.

- `go get github.com/tetratom/qb`
- [GoDoc](https://godoc.org/github.com/tetratom/qb)
- More examples can be found in [qb_test.go](./qb_test.go).
- All methods take value receivers and return values.
- Select the placeholder dialect with the `DialectOption(Dialect)` method.
- Example:

```go
import "github.com/tetratom/qb"

q := qb.
    Select("*").From("my_table").
    Where(qb.
        And("id = ?", 1).
        Or("time < ?", qb.Lit("now()"))).
    OrderBy("time ASC").
    Limit(10)

// q.SQL() is "SELECT * FROM my_table WHERE id = ?".
// q.Args() is []interface{1}.
row := tx.QueryRow(q.SQL(), q.Args()...)
```
