//go:build cgo

package test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMonitorEnumProcCallback(t *testing.T) {
	err := EnumDisplayMonitors()
	assert.NoError(t, err)
}
