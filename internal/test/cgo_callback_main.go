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
	"errors"
	"github.com/jamesits/goinvoke"
	"golang.org/x/sys/windows"
)

var dwData = uintptr(114514)
var callbackDataValidated = true

//export MonitorEnumProcCallback
func MonitorEnumProcCallback(unnamedParam1 C.uintptr_t, unnamedParam2 C.uintptr_t, unnamedParam3 C.uintptr_t, unnamedParam4 C.uintptr_t) C.bool {
	if uintptr(unnamedParam4) != dwData {
		callbackDataValidated = false
		return C.bool(false)
	}

	return C.bool(true)
}

type user32 struct {
	// no tag, use field name to match function
	EnumDisplayMonitors *windows.LazyProc
}

func EnumDisplayMonitors() error {
	var err error

	u := user32{}
	err = goinvoke.Unmarshal("user32.dll", &u)
	if err != nil {
		return err
	}

	_, _, err = u.EnumDisplayMonitors.Call(
		uintptr(0),                            // hdc = NULL
		uintptr(0),                            // lprcClip = NULL
		uintptr(C.monitor_enum_proc_callback), // lpfnEnum
		dwData,                                // dwData
	)

	if err != windows.ERROR_SUCCESS {
		return err
	}

	if !callbackDataValidated {
		return errors.New("dwData validation failed")
	}

	return nil
}
