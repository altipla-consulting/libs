
# errors

[![GoDoc](https://godoc.org/libs.altipla.consulting/errors?status.svg)](https://godoc.org/libs.altipla.consulting/errors)

Package `errors` is a custom library to annotate errors.

Based on https://godoc.org/github.com/samsarahq/go/oops and adapted to our function names and needs.


### Install

```go
import (
	"libs.altipla.consulting/errors"
)
```


### Avoiding manual errors

To reach the maximum help from this library all errors should be annotated. The easiest way to do it is following these rules:
- Ban the import `"errors"` anywhere in the project.
- Ban the import `"github.com/altipla-consulting/errors"` anywhere in the project.
- Ban the import `"github.com/juju/errors"` anywhere in the project.
- Ban the use of `fmt.Errorf` anywhere in the project.
- Ban the use of `errors.New` if it is not a global declaration.
- Ban the use of `return err` and replace it with `return errors.Trace(err)` everywhere.


### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using ```make gofmt```.


### Running tests

Run the tests:

```shell
make test
```


### License

[MIT License](../LICENSE)
