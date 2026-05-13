package mxl

/*
#include <mxl/flowinfo.h>
#include "internal_helpers.h"
*/
import "C"

import "unsafe"

func discreteFromUnion(c *C.mxlFlowConfigInfo) unsafe.Pointer {
	return unsafe.Pointer(C.mxl_go_config_discrete(c))
}

func continuousFromUnion(c *C.mxlFlowConfigInfo) unsafe.Pointer {
	return unsafe.Pointer(C.mxl_go_config_continuous(c))
}
