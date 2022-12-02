# fun
Fun utilities for Golang

Don't use this in production

## pkg

Install:
```bash
go get github.com/itura/fun@v0.1.8
```

Use:

```go
package build

import (
	"github.com/itura/fun"
	"log"
)

const (
	notFound = fun.Error("not found")
)

func main() {
    log.Fatal(notFound)
}

```

## cmd

```bash
go run github.com/itura/fun/cmd/build@v0.1.8
```