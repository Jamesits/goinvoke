//go:build windows

package test

// contains external definitions

import (
	"fmt"
	"github.com/jamesits/goinvoke"
	"golang.org/x/sys/windows"
)

type shcore struct {
	SetProcessDpiAwareness *windows.LazyProc
}

var Shcore = shcore{}

type user32 struct {
	EnumDisplayMonitors           *windows.LazyProc
	SetProcessDPIAware            *windows.LazyProc
	SetProcessDpiAwarenessContext *windows.LazyProc
}

var User32 = user32{}

func init() {
	var err error

	err = goinvoke.Unmarshal("shcore.dll", &Shcore)
	if err != nil {
		panic(err)
	}

	err = goinvoke.Unmarshal("user32.dll", &User32)
	if err != nil {
		panic(err)
	}
}

type ProcessDPIAwareness uintptr

const (
	ProcessDPIUnaware         ProcessDPIAwareness = 0
	ProcessSystemDPIAware     ProcessDPIAwareness = 1
	ProcessPerMonitorDPIAware ProcessDPIAwareness = 2
)

type DPIAwarenessContext uintptr

const (
	DPIAwarenessContextUnaware           DPIAwarenessContext = 16
	DPIAwarenessContextSystemAware       DPIAwarenessContext = 17
	DPIAwarenessContextPerMonitorAware   DPIAwarenessContext = 18
	DPIAwarenessContextPerMonitorAwareV2 DPIAwarenessContext = 34
)

type rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

func (r rect) String() string {
	return fmt.Sprintf("from (%d, %d) to (%d, %d), size %dx%d", r.left, r.top, r.right, r.bottom, r.right-r.left, r.bottom-r.top)
}

type monitor struct {
	hMonitor       uintptr
	hDeviceContext uintptr
	rect           rect
}

func (m monitor) String() string {
	return fmt.Sprintf("hMonitor = 0x%x, hDC = 0x%x, rect = %s", m.hMonitor, m.hDeviceContext, m.rect)
}
