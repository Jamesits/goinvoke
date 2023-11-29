//go:build windows

package benchmark

import (
	"testing"
)

//go:generate "invoker" "-dll" "kernel32.dll" "-generate"
var globalKernel32 Kernel32

func BenchmarkUnmarshalKernel32(b *testing.B) {
	var localKernel32 Kernel32

	for n := 0; n < b.N; n++ {
		localKernel32 = Kernel32{}

		for i := 0; i < len(testSet); i++ {
			_ = localKernel32.Unmarshal("kernel32.dll")
		}
	}

	globalKernel32 = localKernel32
}

//go:generate "invoker" "-dll" "user32.dll" "-generate"
var globalUser32 User32

func BenchmarkUnmarshalUser32(b *testing.B) {
	var localUser32 User32

	for n := 0; n < b.N; n++ {
		localUser32 = User32{}

		for i := 0; i < len(testSet); i++ {
			_ = localUser32.Unmarshal("user32.dll")
		}
	}

	globalUser32 = localUser32
}
