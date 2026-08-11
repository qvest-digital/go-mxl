package mxl

/*
#include <mxl/mxl.h>
*/
import "C"

import (
	"errors"
	"fmt"
)

// Status mirrors the C mxlStatus enum. The zero value (OK) is success; every
// other value implements the error interface, so most callers will see Status
// values through the standard error return.
type Status int32

// statusKinder bridges errors.Is from a Status to sibling-package sentinels
// (e.g. fabrics.ErrNotReady) without importing those packages. Any error
// type that maps to a single libmxl Status implements MxlStatus; Status.Is
// then matches that target when their codes agree.
type statusKinder interface {
	MxlStatus() Status
}

// Is reports whether target identifies the same libmxl status as s. Matches
// targets that expose a MxlStatus() Status method (typically sibling-package
// sentinels) so errors.Is reaches them regardless of which side carries the
// raw Status value.
func (s Status) Is(target error) bool {
	if k, ok := target.(statusKinder); ok {
		return k.MxlStatus() == s
	}
	return false
}

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

// The fabrics half of mxlStatus, numbered from 1024. libmxl-fabrics
// returns these through the same mxlStatus type as the flow API, so a
// caller that only names the flow half sees them as unrecognized
// integers and cannot tell a retryable interruption from a dead
// endpoint.
const (
	StatusErrStrLen               Status = C.MXL_ERR_STRLEN
	StatusErrInterrupted          Status = C.MXL_ERR_INTERRUPTED
	StatusErrNoFabric             Status = C.MXL_ERR_NO_FABRIC
	StatusErrInvalidState         Status = C.MXL_ERR_INVALID_STATE
	StatusErrInternal             Status = C.MXL_ERR_INTERNAL
	StatusErrNotReady             Status = C.MXL_ERR_NOT_READY
	StatusErrNotFound             Status = C.MXL_ERR_NOT_FOUND
	StatusErrExists               Status = C.MXL_ERR_EXISTS
	StatusErrUnsupportedOperation Status = C.MXL_ERR_UNSUPPORTED_OPERATION
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

	// Fabrics half. ErrNotReady keeps its home in the fabrics package,
	// which wraps it in a type that also matches this status.
	ErrStrLen               = StatusErrStrLen
	ErrInterrupted          = StatusErrInterrupted
	ErrNoFabric             = StatusErrNoFabric
	ErrInvalidState         = StatusErrInvalidState
	ErrInternal             = StatusErrInternal
	ErrNotFound             = StatusErrNotFound
	ErrExists               = StatusErrExists
	ErrUnsupportedOperation = StatusErrUnsupportedOperation
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
	case StatusErrStrLen:
		return "mxl: string buffer too small"
	case StatusErrInterrupted:
		return "mxl: interrupted"
	case StatusErrNoFabric:
		return "mxl: no fabric"
	case StatusErrInvalidState:
		return "mxl: invalid state"
	case StatusErrInternal:
		return "mxl: internal error"
	case StatusErrNotReady:
		return "mxl: not ready"
	case StatusErrNotFound:
		return "mxl: not found"
	case StatusErrExists:
		return "mxl: already exists"
	case StatusErrUnsupportedOperation:
		return "mxl: unsupported operation"
	case StatusErrUnknown:
		return "mxl: unknown error"
	default:
		return "mxl: unrecognized status"
	}
}

// known reports whether s is one of the named Status values that Error
// renders with a specific string. Used by StatusErrFromInt32 to route
// unknown codes to UnrecognizedStatusError.
func (s Status) known() bool {
	switch s {
	case StatusOK,
		StatusErrUnknown,
		StatusErrFlowNotFound,
		StatusErrOutOfRangeLate,
		StatusErrOutOfRangeEarly,
		StatusErrInvalidReader,
		StatusErrInvalidWriter,
		StatusErrTimeout,
		StatusErrInvalidArg,
		StatusErrConflict,
		StatusErrPermissionDenied,
		StatusErrFlowInvalid,
		StatusErrStrLen,
		StatusErrInterrupted,
		StatusErrNoFabric,
		StatusErrInvalidState,
		StatusErrInternal,
		StatusErrNotReady,
		StatusErrNotFound,
		StatusErrExists,
		StatusErrUnsupportedOperation:
		return true
	}
	return false
}

// UnrecognizedStatusError is returned by status conversion when libmxl
// reports a status code that this binding does not yet name. Status carries
// the raw integer; Symbol carries a string name when one is known,
// otherwise empty.
type UnrecognizedStatusError struct {
	Status int32
	Symbol string
}

func (e *UnrecognizedStatusError) Error() string {
	if e.Symbol != "" {
		return fmt.Sprintf("mxl: unrecognized status %d (%s)", e.Status, e.Symbol)
	}
	return fmt.Sprintf("mxl: unrecognized status %d", e.Status)
}

// StatusErrFromInt32 converts a raw libmxl status code to a Go error using
// the same rules as the internal statusErr. OK becomes nil; named codes
// return the matching Status; unknown codes return *UnrecognizedStatusError
// carrying the raw integer. Exported so sibling packages that already hold
// the status as int32 share one entry point.
func StatusErrFromInt32(raw int32) error {
	if Status(raw) == StatusOK {
		return nil
	}
	status := Status(raw)
	if !status.known() {
		return &UnrecognizedStatusError{Status: raw}
	}
	return status
}

// statusErr converts a C status code to a Go error. OK becomes nil.
func statusErr(s C.mxlStatus) error {
	return StatusErrFromInt32(int32(s))
}
