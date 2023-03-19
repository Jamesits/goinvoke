//go:build windows

package goinvoke

import (
	"github.com/hashicorp/go-multierror"
	"github.com/jamesits/goinvoke/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

// sanity tests, plus some demonstration of how to use this library
// some unit tests are from https://github.com/dotnet/pinvoke/blob/c01077e0511e6cec6d860b3373ca2a60ba7cbcae/test/Kernel32.Tests/Kernel32Facts.cs
type kernel32 struct {
	GetTickCount              *windows.Proc
	GetTickCount64            *windows.Proc
	SetLastError              *windows.Proc
	SetErrorMode              *windows.Proc
	GetStartupInfoW           *windows.Proc
	GetStartupInfoA           *windows.Proc
	GetSystemInfo             *windows.Proc
	QueryPerformanceFrequency *windows.Proc
	QueryPerformanceCounter   *windows.Proc
}

// value mapping from runtime.GOARCH (plus some non-existent ones) to SYSTEM_INFO.wProcessorArchitecture
// Note: some values are not really useful, they are here for completeness only
// References:
// - https://learn.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-system_info
// - https://learn.microsoft.com/en-us/previous-versions/ms942639(v=msdn.10)
// - Winnt.h
var processorArchitectureMap = map[string]uint16{
	"386":        0,
	"mips":       1,
	"alpha":      2,
	"ppc":        3,
	"shx":        4,
	"arm":        5,
	"ia64":       6,
	"alpha64":    7,
	"msil":       8,
	"amd64":      9,
	"amd64p32":   10,
	"neutral":    11,
	"arm64":      12,
	"arm32win64": 13,
	"ia32arm64":  14,
	"unknown":    0xFFFF, // -1 in int16
}

func processorArchitecture() uint16 {
	ret, ok := processorArchitectureMap[runtime.GOARCH]
	if !ok {
		return processorArchitectureMap["unknown"]
	}

	return ret
}

func TestUnmarshalKernel32(t *testing.T) {
	var ret1, ret2 uintptr
	var err error

	k := kernel32{}
	kernel32Dll := "kernel32.dll" // should only search for system paths (secure mode)

	err = Unmarshal(kernel32Dll, &k)
	assert.NoError(t, err)

	// GetTicketCount should return a non-zero value
	ret1, ret2, err = k.GetTickCount.Call()
	assert.NotZero(t, ret1)
	assert.Zero(t, ret2)
	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)

	// GetTickCount64 should return a non-zero value
	ret1, ret2, err = k.GetTickCount64.Call()
	assert.NotZero(t, ret1)
	assert.Zero(t, ret2)
	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)

	// SetLastError should set the error
	ret1, ret2, err = k.SetLastError.Call(uintptr(windows.ERROR_FILE_NOT_FOUND))
	assert.EqualValues(t, windows.ERROR_FILE_NOT_FOUND, ret1)
	assert.EqualValues(t, windows.ERROR_FILE_NOT_FOUND, ret2)
	assert.ErrorIs(t, err, windows.ERROR_FILE_NOT_FOUND)

	// SetErrorMode should work
	ret1, ret2, err = k.SetErrorMode.Call(uintptr(0))
	assert.NotZero(t, ret1)
	assert.Zero(t, ret2)
	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)

	// GetStartupInfoW should return something from its reference argument
	startupInfoW := windows.StartupInfo{}
	ret1, ret2, err = k.GetStartupInfoW.Call(uintptr(unsafe.Pointer(&startupInfoW)))
	assert.NotZero(t, ret1)
	assert.NotZero(t, ret2)
	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)
	// string conversion: https://gist.github.com/NaniteFactory/9e9d3fe5ea7bfeed788b0795162201c7
	lpTitleW := windows.UTF16PtrToString(startupInfoW.Title)
	assert.True(t, len(lpTitleW) > 0)
	lpDesktopW := windows.UTF16PtrToString(startupInfoW.Desktop)
	assert.EqualValues(t, "Winsta0\\Default", lpDesktopW)

	// GetStartupInfoA does too, but returns ASCII strings
	startupInfoA := windows.StartupInfo{}
	ret1, _, err = k.GetStartupInfoA.Call(uintptr(unsafe.Pointer(&startupInfoA)))
	assert.NotZero(t, ret1)
	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)
	lpTitleA := utils.UintPtrToString(uintptr(unsafe.Pointer(startupInfoA.Title)))
	assert.True(t, len(lpTitleA) > 0)
	lpDesktopA := utils.UintPtrToString(uintptr(unsafe.Pointer(startupInfoA.Desktop)))
	assert.EqualValues(t, "Winsta0\\Default", lpDesktopA)

	// It's not always that you can construct a struct in Golang that matches a certain struct in C in the terms of
	// memory layout. Memory layout of a certain struct might change if your code runs on different architecture as well.
	// Here we demonstrate how to use a byte buffer to hold the returned struct, then decode it with known offsets.
	systemInfo := make([]byte, 64) // size is arbitrary; just make sure it is large enough
	ret1, ret2, err = k.GetSystemInfo.Call(uintptr(unsafe.Pointer(&systemInfo[0])))
	assert.Zero(t, ret1)
	assert.Zero(t, ret2)
	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)
	wProcessorArchitecture := utils.HostByteOrder.Uint16(systemInfo[0:2])
	assert.EqualValues(t, processorArchitecture(), wProcessorArchitecture)
	dwPageSize := utils.HostByteOrder.Uint32(systemInfo[4:8])
	assert.EqualValues(t, os.Getpagesize(), dwPageSize)

	// test QueryPerformanceCounter
	// code adopted from https://stackoverflow.com/a/1739265
	var freq, count1, count2 uint64
	ret1, _, err = k.QueryPerformanceFrequency.Call(uintptr(unsafe.Pointer(&freq)))
	assert.EqualValues(t, 1, ret1) // If the installed hardware supports a high-resolution performance counter, the return value is nonzero.
	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)
	ret1, _, err = k.QueryPerformanceCounter.Call(uintptr(unsafe.Pointer(&count1)))
	assert.EqualValues(t, 1, ret1) // If the function succeeds, the return value is nonzero.
	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)
	time.Sleep(1 * time.Second)
	ret1, ret2, err = k.QueryPerformanceCounter.Call(uintptr(unsafe.Pointer(&count2)))
	assert.EqualValues(t, 1, ret1)
	assert.ErrorIs(t, err, windows.ERROR_SUCCESS)
	assert.LessOrEqual(t, math.Abs(float64(count2-count1)/float64(freq)-1.0), 0.05) // accept a deviation of 5%
}

// unmarshal tests for windows.LazyProc
type user32 struct {
	// no tag, use field name to match function
	MessageBoxA *windows.LazyProc

	// tag overrides field name
	MessageBox *windows.LazyProc `func:"MessageBoxW"`

	// == private fields ===
	// should not be touched in any way
	RandomField int
	gks         *windows.LazyProc `func:"GetKeyState"`
}

func TestUnmarshalUser32(t *testing.T) {
	var err error

	u := user32{
		RandomField: 0,
	}
	assert.Nil(t, u.MessageBoxA)

	sd, err := utils.GetSystemDirectory()
	assert.NoError(t, err)

	user32Dll := filepath.Join(sd, "user32.dll")

	err = Unmarshal(user32Dll, &u)
	assert.NoError(t, err)

	// LazyProc
	assert.NotNil(t, u.MessageBoxA)
	assert.NotNil(t, u.MessageBox)
	assert.NotEqualValues(t, u.MessageBoxA.Addr(), u.MessageBox.Addr())

	// sanity checks
	assert.Zero(t, u.RandomField)
	assert.Nil(t, u.gks)
}

// `ordinal` tag functionality
// test case from: https://github.com/golang/go/issues/16507
// https://learn.microsoft.com/en-us/windows/win32/api/shlwapi/nf-shlwapi-shcreatememstream#remarks
type shlwapi struct {
	// no tag, use field name to match function
	// works only on Windows Vista or later
	SHCreateMemStream *windows.LazyProc

	// only function tag
	SHCreateMemStreamByFunc *windows.Proc `func:"SHCreateMemStream"`

	// you can also import functions by its ordinal
	SHCreateMemStreamByOrdinal *windows.Proc `ordinal:"12"`

	// ordinal, if defined, always takes precedence
	SHCreateMemStreamTestOrdinalOverride  *windows.Proc `ordinal:"12" func:"SHCreateStreamOnFileEx"`
	SHCreateMemStreamTestOrdinalOverride2 *windows.Proc `func:"FunctionThatDoesNotExistAtAll" ordinal:"12"`
}

func TestUnmarshalShlwapi(t *testing.T) {
	var err error

	s := shlwapi{}

	err = Unmarshal("shlwapi.dll", &s)
	assert.NoError(t, err)

	assert.NotNil(t, s.SHCreateMemStream)

	assert.NotNil(t, s.SHCreateMemStreamByFunc)
	assert.EqualValues(t, s.SHCreateMemStream.Addr(), s.SHCreateMemStreamByFunc.Addr())

	assert.NotNil(t, s.SHCreateMemStreamByOrdinal)
	assert.EqualValues(t, s.SHCreateMemStream.Addr(), s.SHCreateMemStreamByOrdinal.Addr())

	assert.NotNil(t, s.SHCreateMemStreamTestOrdinalOverride)
	assert.EqualValues(t, s.SHCreateMemStream.Addr(), s.SHCreateMemStreamTestOrdinalOverride.Addr())
	assert.NotNil(t, s.SHCreateMemStreamTestOrdinalOverride2)
	assert.EqualValues(t, s.SHCreateMemStream.Addr(), s.SHCreateMemStreamTestOrdinalOverride2.Addr())
}

func TestFileMissing(t *testing.T) {
	u := user32{
		RandomField: 0,
	}

	user32Dll := "do_not_exist.dll"

	err := Unmarshal(user32Dll, &u)
	assert.Error(t, err)
	assert.EqualValues(t, 2, len(err.(*multierror.Error).Errors))

	assert.Zero(t, u.RandomField)

	assert.Nil(t, u.MessageBoxA)
	assert.Nil(t, u.MessageBox)
}

func TestProcMissing(t *testing.T) {
	type fakeUser32 struct {
		FunctionMissing1 *windows.LazyProc
		FunctionMissing2 *windows.LazyProc `func:"ArbitraryFakeName"`
	}
	u := fakeUser32{}
	user32Dll := "user32.dll"

	err := Unmarshal(user32Dll, &u)
	assert.Error(t, err)
	assert.EqualValues(t, 3, len(err.(*multierror.Error).Errors))

	assert.Nil(t, u.FunctionMissing1)
	assert.Nil(t, u.FunctionMissing2)
}
