
# content

[![GoDoc](https://godoc.org/libs.altipla.consulting/content?status.svg)](https://godoc.org/libs.altipla.consulting/content)

Package `content` has models for translated content coming from multiple providers.


### Install

```go
import (
	"libs.altipla.consulting/content"
)
```


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


### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using ```make gofmt```.


### Running tests

Run the tests:

```shell
make test
```


### License

[MIT License](../LICENSE)
