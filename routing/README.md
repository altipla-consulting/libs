
# routing

[![GoDoc](https://godoc.org/libs.altipla.consulting/routing?status.svg)](https://godoc.org/libs.altipla.consulting/routing)

Package `routing` sends requests to handlers and process the errors.


### Install

```go
import (
	"libs.altipla.consulting/routing"
)
```


### Basic usage

```go
package main

import (
  "fmt"
  "net/http"

  "libs.altipla.consulting/routing"
  "libs.altipla.consulting/langs"
  "github.com/julienschmidt/httprouter"
)

func RobotsHandler(w http.ResponseWriter, r *http.Request) error {
  fmt.Fprintln(w, "ok")
  return nil
}

func main {
  s := routing.NewServer(r)
  s.Get(langs.ES, "/robots.txt", RobotsHandler)
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
