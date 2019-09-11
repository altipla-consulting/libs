
# redis

[![GoDoc](https://godoc.org/libs.altipla.consulting/redis?status.svg)](https://godoc.org/libs.altipla.consulting/redis)

Package `redis` is an abstraction layer to access Redis with repositories and models.


### Install

```go
import (
	"libs.altipla.consulting/redis"
)
```


### Basic usage

Repository file:

```go
package models

import (
  "fmt"

  "libs.altipla.consulting/redis"
)

var Repo *MainDatabase

func ConnectRepo() error {
  Repo = &MainDatabase{
    sess: redis.Open("redis:6379", "cells"),
  }

  return nil
}

type MainDatabase struct {
  sess *redis.Database
}

func (repo *MainDatabase) Offers(hotel int64) *redis.ProtoHash {
  return repo.sess.ProtoHash(fmt.Sprintf("hotel:%d", hotel))
}
```

App file:

```go
func run() error {
  offers := []*pbmodels.Offer{}
  if err := models.Repo.Offers(in.Hotel).GetMulti(codes, &offers); err != nil {
    return errors.Trace(err)
  }
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
