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
	fmt.Printf("GetTickCount() = %d\n", count)

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

# Caveats

## Type Generator

```shell
# first time
go install github.com/jamesits/goinvoke/cmd/invoker@latest
invoker -dll "kernel32.dll" -generate

# in the future, when your DLL is updated with new exported functions
go generate ./...
```

For detailed usage of `invoker` tool, run `invoker -help`.

## Relative Import

Due to security concerns, if the path is relative and only contains a base name (e.g. `"kernel32.dll"`), file lookup
is limited to *only* Windows system directories. 

If you want to load a DLL packaged with your program, the correct way is to get the program base directory first:
```go
package main

import (
	"github.com/jamesits/goinvoke"
	"os"
	"path/filepath"
)

type MyDll struct {
	// ...
}

func main() {
	executablePath, err := os.Executable()
	if err != nil {
		// process error
		panic(err)
	}
	// directory containing current executable
	executableBaseDir := filepath.Dir(executablePath) 

	myDll := MyDll{}
	err = goinvoke.Unmarshal(filepath.Join(executableBaseDir, "filename.dll"), &myDll)
}
```

If you really want to load a DLL from your *working directory* (bad practice, strongly not recommended!), specify your 
intention explicitly with ".\\filename.dll".
