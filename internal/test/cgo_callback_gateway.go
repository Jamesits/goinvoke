//go:build cgo

package test

/*
#include "common.h"

bool monitor_enum_proc_callback(uintptr_t unnamedParam1, uintptr_t unnamedParam2, uintptr_t unnamedParam3, uintptr_t unnamedParam4) {
	return MonitorEnumProcCallback(unnamedParam1, unnamedParam2, unnamedParam3, unnamedParam4);
}
*/
import "C"
