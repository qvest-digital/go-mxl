package mxl

import (
	"errors"
	"fmt"
	"testing"
)

func TestStatusErrorStrings(t *testing.T) {
	cases := []Status{
		StatusOK, StatusErrUnknown, StatusErrFlowNotFound,
		StatusErrOutOfRangeLate, StatusErrOutOfRangeEarly,
		StatusErrInvalidReader, StatusErrInvalidWriter,
		StatusErrTimeout, StatusErrInvalidArg, StatusErrConflict,
		StatusErrPermissionDenied, StatusErrFlowInvalid,
	}
	for _, s := range cases {
		t.Run(fmt.Sprintf("status=%d", s), func(t *testing.T) {
			if s.Error() == "" {
				t.Fatalf("empty Error() string for status %d", s)
			}
		})
	}
}

func TestStatusErrorUnrecognized(t *testing.T) {
	if got := Status(-999).Error(); got != "mxl: unrecognized status" {
		t.Fatalf("got %q, want %q", got, "mxl: unrecognized status")
	}
}

func TestStatusErrIs(t *testing.T) {
	var err error = StatusErrTimeout
	if !errors.Is(err, ErrTimeout) {
		t.Fatal("errors.Is(StatusErrTimeout, ErrTimeout) = false")
	}
	if errors.Is(err, ErrFlowNotFound) {
		t.Fatal("StatusErrTimeout matched ErrFlowNotFound")
	}
}

func TestStatusErrAs(t *testing.T) {
	var err error = StatusErrFlowInvalid
	var s Status
	if !errors.As(err, &s) {
		t.Fatal("errors.As failed to recover Status")
	}
	if s != StatusErrFlowInvalid {
		t.Fatalf("recovered status %v, want %v", s, StatusErrFlowInvalid)
	}
}

func TestStatusErrNilOnOK(t *testing.T) {
	if err := statusErr(0); err != nil {
		t.Fatalf("statusErr(OK) = %v, want nil", err)
	}
}

func TestErrClosedDistinctFromStatus(t *testing.T) {
	if errors.Is(ErrClosed, StatusErrUnknown) {
		t.Fatal("ErrClosed must not match an arbitrary Status")
	}
	var s Status
	if errors.As(ErrClosed, &s) {
		t.Fatal("ErrClosed must not satisfy errors.As to Status")
	}
}
