//go:build linux

package goinvoke

import (
	"github.com/jamesits/goinvoke/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

type LibC struct {
	Puts *Proc `func:"puts"`
}

var libC LibC

func TestUnmarshal(t *testing.T) {
	err := Unmarshal("libc.so.6", &libC)
	assert.NoError(t, err)
	assert.NotNil(t, libC.Puts)

	ret, _, _ := libC.Puts.Call(utils.StringToUintPtr("114514\n"))
	assert.True(t, ret > 0)
}
