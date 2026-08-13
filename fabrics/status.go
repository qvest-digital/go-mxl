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

// MxlStatus implements the unexported bridge interface in mxl so that
// errors.Is matches ErrNotReady whenever the chain carries a
// mxl.StatusErrNotReady value, regardless of direction.
func (notReadyError) MxlStatus() mxl.Status { return mxl.StatusErrNotReady }

func (notReadyError) Is(target error) bool {
	if _, ok := target.(notReadyError); ok {
		return true
	}
	return errors.Is(target, mxl.StatusErrNotReady)
}

// ErrInterrupted is returned when a POSIX signal interrupts a fabrics call.
// Both the blocking and the non-blocking progress and read calls may report
// it, so it describes an operation that made no progress rather than one that
// failed: the caller retries. Go runtimes send SIGURG to running goroutines
// for async preemption, which makes interruption routine rather than
// exceptional. Matches mxl.StatusErrInterrupted under errors.Is, regardless of
// which return path surfaced the underlying status.
var ErrInterrupted error = interruptedError{}

type interruptedError struct{}

func (interruptedError) Error() string { return "mxl/fabrics: interrupted" }

// MxlStatus implements the unexported bridge interface in mxl so that
// errors.Is matches ErrInterrupted whenever the chain carries a
// mxl.StatusErrInterrupted value, regardless of direction.
func (interruptedError) MxlStatus() mxl.Status { return mxl.StatusErrInterrupted }

func (interruptedError) Is(target error) bool {
	if _, ok := target.(interruptedError); ok {
		return true
	}
	return errors.Is(target, mxl.StatusErrInterrupted)
}

// fabricsStatusErr converts a C mxlStatus into a Go error. MXL_ERR_NOT_READY
// is returned as ErrNotReady and MXL_ERR_INTERRUPTED as ErrInterrupted;
// unknown codes flow through mxl.StatusErrFromInt32 so they surface as
// *mxl.UnrecognizedStatusError carrying the raw integer.
func fabricsStatusErr(rc C.mxlStatus) error {
	if rc == C.MXL_STATUS_OK {
		return nil
	}
	switch mxl.Status(int32(rc)) {
	case mxl.StatusErrNotReady:
		return ErrNotReady
	case mxl.StatusErrInterrupted:
		return ErrInterrupted
	}
	return mxl.StatusErrFromInt32(int32(rc))
}
