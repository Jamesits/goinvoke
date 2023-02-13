//go:build cgo

package benchmark

//#cgo CFLAGS: -I"C:\Program Files (x86)\Windows Kits\10\Include\10.0.19041.0\um"
//#cgo LDFLAGS:

/*
#include <debugapi.h>
*/
import "C"

func IsDebuggerPresent() int {
	return int(C.IsDebuggerPresent())
}
