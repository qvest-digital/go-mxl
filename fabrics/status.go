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
// and related calls when the operation has nothing to report yet. Matches
// mxl.StatusErrNotReady under errors.Is, regardless of which return path
// surfaced the underlying status.
var ErrNotReady error = notReadyError{}

type notReadyError struct{}

func (notReadyError) Error() string { return "mxl/fabrics: not ready" }

func (notReadyError) Is(target error) bool {
	if _, ok := target.(notReadyError); ok {
		return true
	}
	return errors.Is(target, mxl.StatusErrNotReady)
}

// fabricsStatusErr converts a C mxlStatus into a Go error. MXL_ERR_NOT_READY
// is returned as ErrNotReady; unknown codes flow through mxl.StatusErrFromInt32
// so they surface as *mxl.UnrecognizedStatusError carrying the raw integer.
func fabricsStatusErr(rc C.mxlStatus) error {
	if rc == C.MXL_STATUS_OK {
		return nil
	}
	if mxl.Status(int32(rc)) == mxl.StatusErrNotReady {
		return ErrNotReady
	}
	return mxl.StatusErrFromInt32(int32(rc))
}
