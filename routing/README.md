
# routing

[![GoDoc](https://godoc.org/libs.altipla.consulting/routing?status.svg)](https://godoc.org/libs.altipla.consulting/routing)

Routing requests to handlers.


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