package fabrics

/*
#include <mxl/mxl.h>
*/
import "C"

import (
	"errors"

	"github.com/qvest-digital/go-mxl/mxl"
)

// ErrNotReady is returned by ReadGrainNonBlocking, MakeProgressNonBlocking,
// and related calls when the operation has nothing to report yet. Maps
// to MXL_ERR_NOT_READY.
var ErrNotReady = errors.New("mxl/fabrics: not ready")

// fabricsStatusErr converts a C mxlStatus into a Go error. MXL_ERR_NOT_READY
// is returned as ErrNotReady; other errors round-trip through mxl.Status.
func fabricsStatusErr(rc C.mxlStatus) error {
	s := mxl.Status(int32(rc))
	switch s {
	case mxl.StatusOK:
		return nil
	case mxl.Status(C.MXL_ERR_NOT_READY):
		return ErrNotReady
	}
	return s
}
