
# content

[![GoDoc](https://godoc.org/libs.altipla.consulting/content?status.svg)](https://godoc.org/libs.altipla.consulting/content)

Models for translated content coming from multiple providers.


### Basic usage

The types of this package can be used in your models structs when working with `libs.altipla.consulting/database`:

```go
import (
  "libs.altipla.consulting/content"
)

type MyModel struct {
  ID          int64              `db:"id,omitempty"`
  Name        content.Translated `db:"name"`
  Description content.Translated `db:"description"`
  Description content.Provider   `db:"description"`
}
```
