//go:build cgo

package test

// cgo callback documentation:
// - https://eli.thegreenplace.net/2019/passing-callbacks-and-pointers-to-cgo/
// - https://blog.marlin.org/cgo-referencing-c-library-in-go

/*
#include "common.h"
*/
import "C"
import (
	"fmt"
	"github.com/jamesits/goinvoke"
	"github.com/mattn/go-pointer"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows"
	"testing"
	"unsafe"
)

type user32 struct {
	EnumDisplayMonitors *windows.LazyProc
}

type rect struct {
	left   uint32
	top    uint32
	right  uint32
	bottom uint32
}

func (r rect) String() string {
	return fmt.Sprintf("from %dx%d to %dx%d", r.top, r.left, r.bottom, r.right)
}

type monitor struct {
	hMonitor       uintptr
	hDeviceContext uintptr
	rect           rect
}

func (m monitor) String() string {
	return fmt.Sprintf("hMonitor = 0x%x, hDC = 0x%x, rect = %s", m.hMonitor, m.hDeviceContext, m.rect)
}

//export MonitorEnumProcCallback
func MonitorEnumProcCallback(unnamedParam1 C.uintptr_t, unnamedParam2 C.uintptr_t, unnamedParam3 C.uintptr_t, unnamedParam4 C.uintptr_t) C.bool {
	monitors := pointer.Restore(unsafe.Pointer(uintptr(unnamedParam4))).(*[]monitor)

	monitor := monitor{
		hMonitor:       uintptr(unnamedParam1),
		hDeviceContext: uintptr(unnamedParam2),
		rect: rect{ // TODO: fill RECT
			left:   0,
			top:    0,
			right:  0,
			bottom: 0,
		},
	}

	*monitors = append(*monitors, monitor)
	return C.bool(true)
}

func EnumDisplayMonitors(t *testing.T) {
	var err error

	u := user32{}
	err = goinvoke.Unmarshal("user32.dll", &u)
	assert.NoError(t, err)

	monitors := &[]monitor{}

	_, _, err = u.EnumDisplayMonitors.Call(
		uintptr(0),                            // hdc = NULL
		uintptr(0),                            // lprcClip = NULL
		uintptr(C.monitor_enum_proc_callback), // lpfnEnum
		uintptr(pointer.Save(monitors)),       // dwData
	)

	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)
	assert.Greater(t, len(*monitors), 0)
	for i, m := range *monitors {
		fmt.Printf("Monitor #%d: %s\n", i, m)
	}
}
