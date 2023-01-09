package goinvoke

import "C"
import (
	"unsafe"
)

// UintPtrToString converts a raw "const char *" to a Go string.
// https://groups.google.com/g/golang-nuts/c/H77hcVt3AAI
func UintPtrToString(ptr uintptr) string {
	return C.GoString((*C.char)(unsafe.Pointer(ptr)))
}
