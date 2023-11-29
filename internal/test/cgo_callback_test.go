//go:build cgo && windows

package test

import (
	"testing"
)

func TestMonitorEnumProcCallback(t *testing.T) {
	EnumDisplayMonitors(t)
}
