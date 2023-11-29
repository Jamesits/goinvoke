//go:build go1.20

package utils

import "unsafe"

func StringToUintPtr(s string) uintptr {
	return uintptr(unsafe.Pointer(unsafe.StringData(string(append([]byte(s), 0)))))
}
