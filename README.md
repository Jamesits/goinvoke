# goinvoke

Load DLLs and import functions with ease. 

If all you need is an equivalent of `LoadLibrary` and `GetProcAddress` in Go, this library is a lot easier to work with 
than cgo. It does not require a C header to start with, and allows you to dynamically load different DLLs exposing the 
same set of functions.

[![Go Reference](https://pkg.go.dev/badge/github.com/jamesits/goinvoke.svg)](https://pkg.go.dev/github.com/jamesits/goinvoke)

## Usage

Simply define a struct with attributes in the type of `*windows.LazyProc` or `*windows.Proc`, and call 
`goinvoke.Unmarshal("path/to/file.dll", pointerToStruct)`.

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

For more examples of using this library, [`unmarshal_test.go`](unmarshal_test.go) is a good start point. 
[Go: WindowsDLLs](https://github.com/golang/go/wiki/WindowsDLLs) also offers a great view of using 
the `(*windows.LazyProc).Call()` method. If you need to define callback functions, see [`cgo_callback_main.go`](internal/test/cgo_callback_main.go) 
for an example. 

## Type Generator

Have a large DLL with a lot of functions and want to access all of them at once? Use our convenient `invoker` tool to
generate the struct required! For example, if we want to call multiple functions in `user32.dll`, use the following 
commands to generate a "header":
```shell
go install github.com/jamesits/goinvoke/cmd/invoker@latest
invoker -dll "user32.dll" -generate
```

A file named `user32_dll.go` will be generated in the current directory with all the exports from that DLL. To use it 
in your code:
```go
package main

import (
	"github.com/jamesits/goinvoke"
)

func main() {
	var err error
	
	k := User32{}

	// either use the object method
	err = k.Unmarshal("user32.dll")
	// or use the global Unmarshal function
	err = goinvoke.Unmarshal("user32.dll", &k)
	
    // ...
}
```

In the future, when your DLL is updated with new exported functions, just re-generate the file:
```shell
go generate .
```

For advanced usage of this tool, run `invoker -help`.

# Caveats

## Relative Import

Due to security concerns, if the path is relative and only contains a base name (e.g. `"kernel32.dll"`), file lookup
is limited to *only* `%WINDIR%\System32`. 

If you want to load a DLL packaged with your program (the DLL sits right beside your EXE, or under some sub-folder), 
the safe way is to get the directory where your program exists first:
```go
package main

import (
	"github.com/jamesits/goinvoke"
	"github.com/jamesits/goinvoke/utils"
	"path/filepath"
)

type MyDll struct {
	// ...
}

func main() {
	var err error
	
	executableDir, err := utils.ExecutableDir()
	if err != nil {
		panic(err)
	}
	
	myDll := MyDll{}
	err = goinvoke.Unmarshal(filepath.Join(executableDir, /* optional */ "sub-folder", "MyDll.dll"), &myDll)
}
```

If you really want to load a DLL from your *working directory*, specify your intention explicitly 
with `".\\filename.dll"`.
Loading a DLL from an arbitrary working directory might lead to serious security issues. 
DO NOT do this unless you know exactly what you are doing.

## Error Processing

The `Unmarshal()` method returns an error with type `(*multierror.Error)` if any of the following case happens:
- DLL load fails (file does not exist, permission/ACL problem, WDAC/Code Integration policy, etc. )
- The DLL file exists, but a function defined in the struct is not exported by that DLL

It always trys to fill as much as function pointers it can find, and will not be stopped by non-critical errors.
So, depending on your use case, you can ignore certain errors reported by `Unmarshal()`, and use whether the struct 
field is `nil` as an indicator of exported function existence of your loaded DLL file.

If you really want to decode individual errors, use `err.(*multierror.Error).Errors`. There are some examples 
in [`unmarshal_test.go`](unmarshal_test.go).

## Importing Functions by Ordinal

Importing functions by ordinal is fully supported, just use `*windows.Proc` and add a `ordinal` tag. The `ordinal` tag, 
if exists, always overrides the `func` tag.

```go
package main

import (
	"github.com/jamesits/goinvoke"
	"golang.org/x/sys/windows"
)

type shlwapi struct {
	// function definition compatible with Windows XP or earlier
	SHCreateMemStream *windows.Proc `ordinal:"12"`
}

func main() {
	var err error

	s := shlwapi{}
	err = goinvoke.Unmarshal("shlwapi.dll", &s)
	if err != nil {
		panic(err)
	}
	
	// ...
}
```

## Performance

`syscall.Syscall` is somewhat slower due to it allocating heap twice more than a cgo call (variable length arguments, 
and another copy inside `syscall.Syscall()`). There is a `internal/benchmark` package to test the performance of both methods. 
Example result under Go 1.19.1:

```text
goos: windows
goarch: amd64
pkg: github.com/jamesits/goinvoke/internal/benchmark
cpu: AMD Ryzen 9 5900X 12-Core Processor
BenchmarkSyscallIsDebuggerPresent
BenchmarkSyscallIsDebuggerPresent-24            29967184                40.49 ns/op
BenchmarkCgoIsDebuggerPresent
BenchmarkCgoIsDebuggerPresent-24                40003599                29.06 ns/op
PASS
```
