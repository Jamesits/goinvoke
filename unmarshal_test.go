//go:build windows

package goinvoke

import (
	"github.com/hashicorp/go-multierror"
	"github.com/jamesits/goinvoke/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"
)

// sanity tests, plus some demonstration of how to use this library
// some unit tests are from https://github.com/dotnet/pinvoke/blob/c01077e0511e6cec6d860b3373ca2a60ba7cbcae/test/Kernel32.Tests/Kernel32Facts.cs
type kernel32 struct {
	GetTickCount    *windows.LazyProc
	GetTickCount64  *windows.LazyProc
	SetLastError    *windows.LazyProc
	SetErrorMode    *windows.LazyProc
	GetStartupInfoW *windows.LazyProc
	GetStartupInfoA *windows.LazyProc
	GetSystemInfo   *windows.LazyProc
}

// value mapping from runtime.GOARCH to LPSYSTEM_INFO.wProcessorArchitecture
// https://learn.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-system_info
var processorArchitectureMap = map[string]uint16{
	"386":   0,
	"arm":   5,
	"ia64":  6, // not really useful, for completeness only
	"amd64": 9,
	"arm64": 12,
}

func processorArchitecture() uint16 {
	ret, ok := processorArchitectureMap[runtime.GOARCH]
	if !ok {
		return 0xFFFF
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

	// should use secure search order
	assert.True(t, lazyDLLReferenceCache[kernel32Dll].System)

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

	// should use insecure search order since we specified an absolute path
	assert.False(t, lazyDLLReferenceCache[user32Dll].System)

	// LazyProc
	assert.NotNil(t, u.MessageBoxA)
	assert.NotNil(t, u.MessageBox)
	assert.NotEqualValues(t, u.MessageBoxA.Addr(), u.MessageBox.Addr())

	// sanity checks
	assert.Zero(t, u.RandomField)
	assert.Nil(t, u.gks)
}

// unmarshal tests for windows.Proc
// test data from: https://github.com/golang/go/issues/16507
// https://learn.microsoft.com/en-us/windows/win32/api/shlwapi/nf-shlwapi-shcreatememstream#remarks
type shlwapi struct {
	// no tag, use field name to match function
	// works only on Windows Vista or later
	SHCreateMemStream *windows.Proc

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
