
# geo

[![GoDoc](https://godoc.org/libs.altipla.consulting/geo?status.svg)](https://godoc.org/libs.altipla.consulting/geo)

Customized types and functions for our geo needs.

**NOTE:** If you want a full-fledged geo library we recommend using https://github.com/twpayne/go-geom instead.


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
