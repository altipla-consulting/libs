
# arrays

[![GoDoc](https://godoc.org/libs.altipla.consulting/arrays?status.svg)](https://godoc.org/libs.altipla.consulting/arrays)

Models for integer and string arrays in MySQL.


### Basic usage

You can use the types of this package in your models structs when working with `database/sql`, `upper.io/db.v3` or `libs.altipla.consulting/database`:

```go
type MyModel struct {
  ID    int64             `db:"id,omitempty"`
  Foo   arrays.Integers32 `db:"foo"`
  Bar   arrays.Integers64 `db:"bar"`
  Codes arrays.Strings    `db:"codes"`
}
```
