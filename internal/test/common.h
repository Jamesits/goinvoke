#pragma once
#ifndef __COMMON_H__
#define __COMMON_H__

#include <stdint.h>
#include <stdbool.h>

struct rect {
	int32_t left;
	int32_t top;
	int32_t right;
	int32_t bottom;
};

// Golang exported function
extern bool MonitorEnumProcCallback(uintptr_t, uintptr_t, uintptr_t, uintptr_t);

#endif
