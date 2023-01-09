package goinvoke

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"testing"
	"unsafe"
)

// unmarshal tests
type user32 struct {
	// field that should not be touched
	RandomField int

	// no tag, use field name to match function
	MessageBoxA *windows.LazyProc

	// tag overrides field name
	MessageBox *windows.LazyProc `func:"MessageBoxW"`
}

func TestUnmarshalUser32(t *testing.T) {
	u := user32{
		RandomField: 0,
	}
	assert.Nil(t, u.MessageBoxA)

	user32Dll := filepath.Join(os.Getenv("WINDIR"), "System32", "user32.dll")

	err := Unmarshal(user32Dll, &u)
	assert.NoError(t, err)

	// should use insecure search order since we specified an absolute path
	assert.False(t, globalDllReferenceCache[user32Dll].System)

	assert.Zero(t, u.RandomField)

	assert.NotNil(t, u.MessageBoxA)
	assert.NotNil(t, u.MessageBox)
	assert.NotEqualValues(t, u.MessageBoxA, u.MessageBox)
}

// sanity tests
// some unit tests are from https://github.com/dotnet/pinvoke/blob/c01077e0511e6cec6d860b3373ca2a60ba7cbcae/test/Kernel32.Tests/Kernel32Facts.cs
type kernel32 struct {
	GetTickCount   *windows.LazyProc
	GetTickCount64 *windows.LazyProc
	SetLastError   *windows.LazyProc
	SetErrorMode   *windows.LazyProc
	GetStartupInfo *windows.LazyProc `func:"GetStartupInfoW"`
}

func TestUnmarshalKernel32(t *testing.T) {
	var ret1, ret2 uintptr
	var err error

	k := kernel32{}
	kernel32Dll := "kernel32.dll" // should only search for system paths (secure mode)

	err = Unmarshal(kernel32Dll, &k)
	assert.NoError(t, err)

	// should use secure search order
	assert.True(t, globalDllReferenceCache[kernel32Dll].System)

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
	startupInfo := windows.StartupInfo{}
	ret1, ret2, err = k.GetStartupInfo.Call(uintptr(unsafe.Pointer(&startupInfo)))
	// string conversion: https://gist.github.com/NaniteFactory/9e9d3fe5ea7bfeed788b0795162201c7
	lpTitle := windows.UTF16PtrToString(startupInfo.Title)
	assert.True(t, len(lpTitle) > 0)
	lpDesktop := windows.UTF16PtrToString(startupInfo.Desktop)
	assert.EqualValues(t, "Winsta0\\Default", lpDesktop)
}
