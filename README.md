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

- `go get -u github.com/tetratom/qb`
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

# thread-safe and reusable builders

_qb_ builders should be used the same as the `append()` built-in. The builders
take and return values, and internally keep state such that one builder value
can be re-used for different queries.  For example:

```go
qbase := qb.Select("col1", "col2")
q1 := qbase.From("t1") // q1 is: SELECT col1, col2 FROM t1
q2 := qbase.From("t2") // q2 is: SELECT col1, col2 FROM t2
// qbase is: SELECT col1, col2
``` 

Just like with `append()`, the return value of calling a builder method must be
assigned back into a variable to be used. For example:

```go
func Search(name string, ordered bool) qb.Query {
    q := qb.Select("*").From("members")

    if name != "" {
        q = q.Where(qb.And("name = ?", name))
    }
    
    if ordered {
        q = q.OrderBy("created_at DESC")
    }
    
    return q
}
```
