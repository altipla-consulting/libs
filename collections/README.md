
<p align="center">
  <img src="https://storage.googleapis.com/altipla-external-files/logos/collections.png">
</p>
<br>

[![GoDoc](https://godoc.org/libs.altipla.consulting/collections?status.svg)](https://godoc.org/libs.altipla.consulting/collections)

Set of functions that help us work with slices and maps.


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
