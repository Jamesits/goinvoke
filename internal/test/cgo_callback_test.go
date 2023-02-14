//go:build cgo

package test

import (
	"testing"
)

func TestMonitorEnumProcCallback(t *testing.T) {
	EnumDisplayMonitors(t)
}
