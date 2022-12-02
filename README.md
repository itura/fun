# fungo
Fun utilities for Golang

Don't use this in production

Install:
```bash
go get github.com/itura/fun@v0.1.7
```

Use:

```go
package build

import (
	"github.com/itura/fungo"
	"log"
)

const (
	notFound = fun.Error("not found")
)

func main() {
    log.Fatal(notFound)
}

```