#pragma once
#ifndef __COMMON_H__
#define __COMMON_H__

#include <stdint.h>
#include <stdbool.h>

// C gateway function
extern bool monitor_enum_proc_callback(uintptr_t, uintptr_t, uintptr_t, uintptr_t);

// Golang exported function
extern bool MonitorEnumProcCallback(uintptr_t, uintptr_t, uintptr_t, uintptr_t);

#endif
