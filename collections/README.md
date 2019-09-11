
<p align="center">
  <img src="https://storage.googleapis.com/altipla-external-files/logos/collections.png">
</p>
<br>


[![GoDoc](https://godoc.org/libs.altipla.consulting/collections?status.svg)](https://godoc.org/libs.altipla.consulting/collections)

Package `collections` is a set of functions that help us to work with slices and maps.


### Install

```go
import (
	"libs.altipla.consulting/collections"
)
```


### Basic usage

```go
package main

import (
  "fmt"
  
  "libs.altipla.consulting/collections"
)

  func main() {
    goFounders := []string{"Robert Griesemer", "Rob Pike", "Ken Thompson"}
    fmt.Println("Francis McCabe:", collections.HasString(goFounders, "Francis McCabe"))
    fmt.Println("RobertGriesemer:", collections.HasString(goFounders, "RobertGriesemer"))
    fmt.Println("Robert Griesemer:", collections.HasString(goFounders, "Robert Griesemer"))
  }
)
```

Result:
```
Fracis McCabe: false 
RobertGriesemer: false 
Robert Griesemer: true 
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
