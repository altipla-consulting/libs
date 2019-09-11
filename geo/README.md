
# geo

[![GoDoc](https://godoc.org/libs.altipla.consulting/geo?status.svg)](https://godoc.org/libs.altipla.consulting/geo)

Package `geo` implements customized types and functions for our geo needs.

**NOTE:** If you want a full-fledged geo library we recommend using https://github.com/twpayne/go-geom instead.


### Install

```go
import (
	"libs.altipla.consulting/geo"
)
```


### Basic usage

You can use the types of this package in your models structs when working with `libs.altipla.consulting/database`:

```go
import (
  "libs.altipla.consulting/geo"
)

type MyModel struct {
  ID          int64      `db:"id,pk"`
  Location    geo.Point `db:"location"`
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
