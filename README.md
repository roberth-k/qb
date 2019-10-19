<h1 align="center">github.com/tetratom/qb</h1>
<p align="center">
  <a href="https://godoc.org/github.com/tetratom/qb">
    <img src="https://godoc.org/github.com/tetratom/qb?status.svg" alt="GoDoc">
  </a>
  <a href="https://circleci.com/gh/tetratom/qb">
    <img src="https://img.shields.io/circleci/build/gh/tetratom/qb/master" alt="CircleCI">
  </a>
  <a href="https://codecov.io/gh/tetratom/qb">
    <img src="https://img.shields.io/codecov/c/github/tetratom/qb/master" alt="Codecov">
  </a>
</p>
<p align="center">
    qb is a simple SQL query builder for Go
</p>

# highlights

- `go get github.com/tetratom/qb`
- [GoDoc](https://godoc.org/github.com/tetratom/qb)
- More examples can be found in [qb_test.go](./qb_test.go).
- All methods take value receivers and return values.
- Select the placeholder dialect with the `DialectOption(Dialect)` method.

# example

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
