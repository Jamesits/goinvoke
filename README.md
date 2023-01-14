# goinvoke

Load DLLs and import functions with ease. If all you need is `LoadLibrary` and `GetProcAddress`, this library is a lot easier to work with than CGO.

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

	// you can override the function name with a tag
	GetStartupInfo *windows.LazyProc `func:"GetStartupInfoW"`
}

func main() {
	k := Kernel32{}
	err := goinvoke.Unmarshal("kernel32.dll", &k)
	if err != nil {
		panic(err)
	}

	// a minimal example
	count, _, err := k.GetTickCount.Call()
	if !errors.Is(err, windows.ERROR_SUCCESS) {
		panic(err)
	}
	fmt.Printf("GetTickCount() returns %d\n", count)

	// a more complete example
	startupInfo := windows.StartupInfo{}
	_, _, err = k.GetStartupInfo.Call(uintptr(unsafe.Pointer(&startupInfo)))
	if !errors.Is(err, windows.ERROR_SUCCESS) {
		panic(err)
	}
	lpTitle := windows.UTF16PtrToString(startupInfo.Title)
	fmt.Printf("lpTitle = %s\n", lpTitle)
}

```

See [Godoc](https://pkg.go.dev/github.com/jamesits/goinvoke) for detailed documentation.
