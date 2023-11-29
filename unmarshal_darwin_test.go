//go:build darwin

package goinvoke

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"unsafe"
)

type LibC struct {
	Puts *Proc `func:"puts"`
}

var libC LibC

func TestUnmarshal(t *testing.T) {
	err := Unmarshal("/usr/lib/libSystem.B.dylib", &libC)
	assert.NoError(t, err)
	assert.NotNil(t, libC.Puts)

	ret, _, err := libC.Puts.Call(uintptr(unsafe.Pointer(unsafe.StringData(string([]byte{'1', '1', '4', '5', '1', '4', '\n', 0})))))
	assert.NoError(t, err)
	assert.EqualValues(t, uintptr(0xa), ret)
}
