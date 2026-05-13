package mxl

/*
#include <mxl/mxl.h>
*/
import "C"

import "errors"

// Status mirrors the C mxlStatus enum. The zero value (OK) is success; every
// other value implements the error interface, so most callers will see Status
// values through the standard error return.
type Status int32

const (
	StatusOK                  Status = C.MXL_STATUS_OK
	StatusErrUnknown          Status = C.MXL_ERR_UNKNOWN
	StatusErrFlowNotFound     Status = C.MXL_ERR_FLOW_NOT_FOUND
	StatusErrOutOfRangeLate   Status = C.MXL_ERR_OUT_OF_RANGE_TOO_LATE
	StatusErrOutOfRangeEarly  Status = C.MXL_ERR_OUT_OF_RANGE_TOO_EARLY
	StatusErrInvalidReader    Status = C.MXL_ERR_INVALID_FLOW_READER
	StatusErrInvalidWriter    Status = C.MXL_ERR_INVALID_FLOW_WRITER
	StatusErrTimeout          Status = C.MXL_ERR_TIMEOUT
	StatusErrInvalidArg       Status = C.MXL_ERR_INVALID_ARG
	StatusErrConflict         Status = C.MXL_ERR_CONFLICT
	StatusErrPermissionDenied Status = C.MXL_ERR_PERMISSION_DENIED
	StatusErrFlowInvalid      Status = C.MXL_ERR_FLOW_INVALID
)

// Sentinel errors. Compare with errors.Is. The Status values themselves also
// satisfy error and round-trip via errors.As(&status).
var (
	ErrUnknown          = StatusErrUnknown
	ErrFlowNotFound     = StatusErrFlowNotFound
	ErrOutOfRangeLate   = StatusErrOutOfRangeLate
	ErrOutOfRangeEarly  = StatusErrOutOfRangeEarly
	ErrInvalidReader    = StatusErrInvalidReader
	ErrInvalidWriter    = StatusErrInvalidWriter
	ErrTimeout          = StatusErrTimeout
	ErrInvalidArg       = StatusErrInvalidArg
	ErrConflict         = StatusErrConflict
	ErrPermissionDenied = StatusErrPermissionDenied
	ErrFlowInvalid      = StatusErrFlowInvalid
)

// ErrClosed is returned by methods called on a handle after Close.
var ErrClosed = errors.New("mxl: handle closed")

func (s Status) Error() string {
	switch s {
	case StatusOK:
		return "mxl: ok"
	case StatusErrFlowNotFound:
		return "mxl: flow not found"
	case StatusErrOutOfRangeLate:
		return "mxl: out of range (too late)"
	case StatusErrOutOfRangeEarly:
		return "mxl: out of range (too early)"
	case StatusErrInvalidReader:
		return "mxl: invalid flow reader"
	case StatusErrInvalidWriter:
		return "mxl: invalid flow writer"
	case StatusErrTimeout:
		return "mxl: timeout"
	case StatusErrInvalidArg:
		return "mxl: invalid argument"
	case StatusErrConflict:
		return "mxl: conflict"
	case StatusErrPermissionDenied:
		return "mxl: permission denied"
	case StatusErrFlowInvalid:
		return "mxl: flow invalid (replaced by writer)"
	case StatusErrUnknown:
		return "mxl: unknown error"
	default:
		return "mxl: unrecognized status"
	}
}

// statusErr converts a C status code to a Go error. OK becomes nil.
func statusErr(s C.mxlStatus) error {
	if s == C.MXL_STATUS_OK {
		return nil
	}
	return Status(s)
}
