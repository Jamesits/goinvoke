//go:build cgo && windows

package test

// Demonstrates how to write C callbacks in Go.
//
// cgo callback documentation:
// - https://eli.thegreenplace.net/2019/passing-callbacks-and-pointers-to-cgo/
// - https://blog.marlin.org/cgo-referencing-c-library-in-go

/*
#include "common.h"
*/
import "C"
import (
	"fmt"
	"github.com/mattn/go-pointer"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows"
	"reflect"
	"testing"
	"unsafe"
)

// MonitorEnumProcCallback is our callback function which will be called from the DLL.
// Cgo note:
// Using //export in a file places a restriction on the preamble: since it is copied into two different C output files,
// it must not contain any definitions, only declarations. If a file contains both definitions and declarations, then
// the two output files will produce duplicate symbols and the linker will fail. To avoid this, definitions must be
// placed in preambles in other files, or in C source files.
// References:
// - https://groups.google.com/g/golang-nuts/c/yaP27124ly8/m/qiajGsLEBAAJ
// - https://pkg.go.dev/cmd/cgo#hdr-C_references_to_Go
//
//export MonitorEnumProcCallback
func MonitorEnumProcCallback(hMonitor C.uintptr_t, hDC C.uintptr_t, lpRect C.uintptr_t, dwData C.uintptr_t) C.bool {
	monitors := pointer.Restore(unsafe.Pointer(uintptr(dwData))).(*[]monitor)

	// converts a C pointer to an array into a Golang slice
	// References:
	// - https://github.com/golang/go/issues/41705
	// - https://pkg.go.dev/unsafe#Pointer
	// - https://utcc.utoronto.ca/~cks/space/blog/programming/GoMemoryToStructures
	var rectSlice []int32
	rectSliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&rectSlice))
	rectSliceHeader.Cap = 4
	rectSliceHeader.Len = 4
	rectSliceHeader.Data = uintptr(lpRect)

	// converts a C pointer to a cgo struct pointer
	// the struct definition is in "common.h"
	// References:
	// - https://utcc.utoronto.ca/~cks/space/blog/programming/GoCGoCompatibleStructs
	rectCStruct := (*C.struct_rect)(unsafe.Pointer(uintptr(lpRect)))

	monitor := monitor{
		hMonitor:       uintptr(hMonitor),
		hDeviceContext: uintptr(hDC),
		rect: rect{
			// you can access values by pure pointer algorithm
			left: *(*int32)(unsafe.Pointer(uintptr(lpRect))),
			top:  *(*int32)(unsafe.Pointer(uintptr(lpRect) + 4)),
			// or cast the pointer into a Golang slice
			right: rectSlice[2],
			// or convert it into a C struct and access its member
			bottom: int32(rectCStruct.bottom),
		},
	}

	*monitors = append(*monitors, monitor)
	return C.bool(true)
}

// SetProcessDpiAwareness enables per-monitor DPI awareness V2. If successful, returns true; otherwise returns false.
// References:
// - https://stackoverflow.com/a/43537991
// - https://github.com/anaisbetts/PerMonitorDpi/blob/63570f78d9a3ff032bcd0d8a50169af1a57c2090/SafeNativeMethods.cs
func SetProcessDpiAwareness() bool {
	// https://stackoverflow.com/a/75074215
	major, minor, revision := windows.RtlGetNtVersionNumbers()

	if major >= 10 && minor >= 0 && revision >= 15063 {
		// per-monitor DPI awareness V2
		ret, _, _ := User32.SetProcessDpiAwarenessContext.Call(uintptr(DPIAwarenessContextPerMonitorAwareV2))
		return ret != 0 // function returns bool
	} else if major >= 6 && minor >= 3 && revision >= 0 {
		// per-monitor DPI awareness
		ret, _, _ := Shcore.SetProcessDpiAwareness.Call(uintptr(ProcessPerMonitorDPIAware))
		return ret == 0 // function returns HRESULT
	} else {
		ret, _, _ := User32.SetProcessDPIAware.Call()
		return ret != 0 // function returns bool
	}
}

func EnumDisplayMonitors(t *testing.T) {
	var err error

	// set DPI awareness (again)
	// (It may already be set in the application side-by-side configuration, so it might return false)
	_ = SetProcessDpiAwareness()

	// Have to use a pointer here, because we are going to append to the slice in the callback function, and slice
	// address might change
	// References:
	// - https://www.tugberkugurlu.com/archive/working-with-slices-in-go-golang-understanding-how-append-copy-and-slicing-syntax-work#how-append-and-copy-works
	// - https://stackoverflow.com/questions/54195834/how-to-inspect-slice-header/54196005
	monitors := &[]monitor{}

	_, _, err = User32.EnumDisplayMonitors.Call(
		uintptr(0),                         // hdc = NULL
		uintptr(0),                         // lprcClip = NULL
		uintptr(C.MonitorEnumProcCallback), // lpfnEnum
		uintptr(pointer.Save(monitors)),    // dwData
	)

	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)
	assert.Greater(t, len(*monitors), 0)
	for i, m := range *monitors {
		fmt.Printf("Monitor #%d: %s\n", i, m)
	}
}
