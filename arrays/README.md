
# arrays

[![GoDoc](https://godoc.org/libs.altipla.consulting/arrays?status.svg)](https://godoc.org/libs.altipla.consulting/arrays)

Package `arrays` has models for integer and string arrays in MySQL.


### Install

```go
import (
	"libs.altipla.consulting/arrays"
)
```


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


### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using ```make gofmt```.


### Running tests

Run the tests:

```shell
make test
```


### License

[MIT License](../LICENSE)
