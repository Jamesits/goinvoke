package utils

import "C"
import (
	"unsafe"
)

// UintPtrToString converts a zero-terminated raw "const char *" to a Go string.
// Example:
//
//	var str := utils.UintPtrToString(uintptr(unsafe.Pointer(uint16ptr_variable)))
func UintPtrToString(ptr uintptr) string {
	// https://groups.google.com/g/golang-nuts/c/H77hcVt3AAI
	return C.GoString((*C.char)(unsafe.Pointer(ptr)))
}
