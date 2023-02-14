//go:build cgo

package benchmark

// performance comparison: cgo vs (*windows.LazyProc).Call()

import (
	"github.com/jamesits/goinvoke"
	"golang.org/x/sys/windows"
	"testing"
)

type kernel32 struct {
	IsDebuggerPresent     *windows.Proc     `func:"IsDebuggerPresent"`
	IsDebuggerPresentLazy *windows.LazyProc `func:"IsDebuggerPresent"`
}

var isDebuggerPresent uintptr

func BenchmarkSyscallIsDebuggerPresent(b *testing.B) {
	var ret1 uintptr
	var err error

	k := kernel32{}
	kernel32Dll := "kernel32.dll" // should only search for system paths (secure mode)

	err = goinvoke.Unmarshal(kernel32Dll, &k)
	if err != nil {
		b.Fail()
	}

	for n := 0; n < b.N; n++ {
		ret1, _, err = k.IsDebuggerPresent.Call()
	}

	isDebuggerPresent = ret1
}

func BenchmarkSyscallIsDebuggerPresentLazy(b *testing.B) {
	var ret1 uintptr
	var err error

	k := kernel32{}
	kernel32Dll := "kernel32.dll" // should only search for system paths (secure mode)

	err = goinvoke.Unmarshal(kernel32Dll, &k)
	if err != nil {
		b.Fail()
	}

	for n := 0; n < b.N; n++ {
		ret1, _, err = k.IsDebuggerPresentLazy.Call()
	}

	isDebuggerPresent = ret1
}

func BenchmarkCgoIsDebuggerPresent(b *testing.B) {
	var idp int

	for n := 0; n < b.N; n++ {
		idp = IsDebuggerPresent()
	}

	isDebuggerPresent = uintptr(idp)
}
