
# redis

[![GoDoc](https://godoc.org/libs.altipla.consulting/redis?status.svg)](https://godoc.org/libs.altipla.consulting/redis)

Abstraction layer to access Redis with repositories and models.


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
    return err
  }
}
```
