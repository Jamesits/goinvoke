# goinvoke

Load DLLs and import functions with ease.

[![Go Reference](https://pkg.go.dev/badge/github.com/jamesits/goinvoke.svg)](https://pkg.go.dev/github.com/jamesits/goinvoke)

## Usage

```go
package main

import (
	"errors"
	"fmt"
	"github.com/jamesits/goinvoke"
	"golang.org/x/sys/windows"
)

type Kernel32 struct {
	GetTickCount *windows.LazyProc
}

func main() {
	k := Kernel32{}
	err := goinvoke.Unmarshal("kernel32.dll", &k)
	if err != nil {
		panic(err)
	}

	// GetTicketCount should return a non-zero value
	count, _, err := k.GetTickCount.Call()
	if !errors.Is(err, windows.ERROR_SUCCESS) {
		panic(err)
    }
	fmt.Printf("GetTickCount() returns %d\n", count)
}
```

See [Godoc](https://pkg.go.dev/github.com/jamesits/goinvoke) for detailed documentation.
