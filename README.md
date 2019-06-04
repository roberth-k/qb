qb
===

qb is a SQL query builder for Golang. It will never have flashy features.

- `go get github.com/tetratom/qb`
- [GoDoc](https://godoc.org/github.com/tetratom/qb)
- More examples can be found in [qb_test.go](./qb_test.go).
- Example:

```go
import "github.com/tetratom/qb"

q := qb.Select("*").From("my_table").Where(qb.AndS("id = ?", 1))
// q.SQL() is "SELECT * FROM my_table WHERE id = ?".
// q.Args() is []interface{1}.
row := tx.QueryRow(q.SQL(), q.Args()...)
```
