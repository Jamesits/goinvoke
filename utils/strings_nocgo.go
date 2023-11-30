//go:build !cgo

package utils

import "C"
import "unsafe"

// UintPtrToString converts a zero-terminated raw "const char *" to a Go string.
// Example:
//
//	var str := utils.UintPtrToString(uintptr(unsafe.Pointer(uint16ptr_variable)))
//
//go:uintptrescapes
func UintPtrToString(ptr uintptr) string {
	i := uintptr(0)
	for ; ; i++ {
		if byte(*unsafe.Pointer(ptr + i)) == 0 {
			break
		}
	}
	return unsafe.String(ptr, i)
}
