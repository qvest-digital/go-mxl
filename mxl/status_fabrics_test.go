package mxl

import (
	"errors"
	"fmt"
	"testing"
)

// fabricsStatuses is every code in the fabrics half of mxlStatus paired
// with the value libmxl assigns it. The numbers are pinned rather than
// read back from the constants: callers branch on these to tell a
// retryable interruption from a dead endpoint, so a renumbering
// upstream has to fail here rather than silently reclassify errors in
// every consumer.
var fabricsStatuses = []struct {
	status Status
	code   int32
	name   string
}{
	{StatusErrStrLen, 1024, "MXL_ERR_STRLEN"},
	{StatusErrInterrupted, 1025, "MXL_ERR_INTERRUPTED"},
	{StatusErrNoFabric, 1026, "MXL_ERR_NO_FABRIC"},
	{StatusErrInvalidState, 1027, "MXL_ERR_INVALID_STATE"},
	{StatusErrInternal, 1028, "MXL_ERR_INTERNAL"},
	{StatusErrNotReady, 1029, "MXL_ERR_NOT_READY"},
	{StatusErrNotFound, 1030, "MXL_ERR_NOT_FOUND"},
	{StatusErrExists, 1031, "MXL_ERR_EXISTS"},
	{StatusErrUnsupportedOperation, 1032, "MXL_ERR_UNSUPPORTED_OPERATION"},
}

func TestFabricsStatusValuesMatchLibmxl(t *testing.T) {
	for _, tc := range fabricsStatuses {
		t.Run(tc.name, func(t *testing.T) {
			if int32(tc.status) != tc.code {
				t.Fatalf("%s = %d, want %d", tc.name, int32(tc.status), tc.code)
			}
		})
	}
}

func TestFabricsStatusIsNamed(t *testing.T) {
	// A fabrics code arriving as *UnrecognizedStatusError is the defect
	// this mapping exists to close: the caller then has only a raw
	// integer in a log line and cannot branch on the failure at all.
	for _, tc := range fabricsStatuses {
		t.Run(tc.name, func(t *testing.T) {
			err := StatusErrFromInt32(tc.code)
			var ue *UnrecognizedStatusError
			if errors.As(err, &ue) {
				t.Fatalf("%s (%d) surfaced as unrecognized: %v", tc.name, tc.code, err)
			}
			var got Status
			if !errors.As(err, &got) {
				t.Fatalf("%s did not convert to a Status: %v", tc.name, err)
			}
			if got != tc.status {
				t.Fatalf("converted to %d, want %d", int32(got), tc.code)
			}
		})
	}
}

func TestFabricsStatusErrorStrings(t *testing.T) {
	// Every fabrics code has to render as something an operator can read
	// in a gateway log without consulting the header.
	seen := map[string]Status{}
	for _, tc := range fabricsStatuses {
		msg := tc.status.Error()
		if msg == "" || msg == "mxl: unrecognized status" {
			t.Fatalf("%s renders as %q", tc.name, msg)
		}
		if prev, dup := seen[msg]; dup {
			t.Fatalf("%s and status %d share the message %q",
				tc.name, int32(prev), msg)
		}
		seen[msg] = tc.status
	}
}

func TestFabricsStatusSentinelsMatch(t *testing.T) {
	sentinels := []struct {
		sentinel error
		status   Status
	}{
		{ErrStrLen, StatusErrStrLen},
		{ErrInterrupted, StatusErrInterrupted},
		{ErrNoFabric, StatusErrNoFabric},
		{ErrInvalidState, StatusErrInvalidState},
		{ErrInternal, StatusErrInternal},
		{ErrNotFound, StatusErrNotFound},
		{ErrExists, StatusErrExists},
		{ErrUnsupportedOperation, StatusErrUnsupportedOperation},
	}
	for _, tc := range sentinels {
		t.Run(fmt.Sprintf("status=%d", int32(tc.status)), func(t *testing.T) {
			var err error = tc.status
			if !errors.Is(err, tc.sentinel) {
				t.Fatalf("errors.Is(%d, sentinel) = false", int32(tc.status))
			}
			if errors.Is(err, ErrTimeout) {
				t.Fatalf("status %d matched an unrelated sentinel", int32(tc.status))
			}
		})
	}
}

func TestUnrecognizedStillReachableOutsideTheEnum(t *testing.T) {
	// The fabrics range is contiguous and ends at
	// MXL_ERR_UNSUPPORTED_OPERATION. A code past it is a libmxl this
	// binding has not caught up with, and must stay distinguishable
	// from one it has named.
	err := StatusErrFromInt32(int32(StatusErrUnsupportedOperation) + 1)
	var ue *UnrecognizedStatusError
	if !errors.As(err, &ue) {
		t.Fatalf("code past the enum did not surface as unrecognized: %v", err)
	}
}
