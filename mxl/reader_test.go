package mxl

import (
	"errors"
	"testing"
	"time"
)

func newTestReader(t *testing.T, inst *Instance, flowID string) *Reader {
	t.Helper()
	r, err := inst.NewReader(flowID)
	if err != nil {
		t.Fatalf("NewReader(%s): %v", flowID, err)
	}
	t.Cleanup(func() { r.Close() })
	return r
}

func TestNewReaderMissingFlow(t *testing.T) {
	inst := newTestInstance(t)
	_, err := inst.NewReader("00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrFlowNotFound) {
		t.Fatalf("NewReader(missing flow): %v, want ErrFlowNotFound", err)
	}
}

func TestReaderCloseIdempotent(t *testing.T) {
	inst := newTestInstance(t)
	_ = newTestWriter(t, inst, testVideoFlowJSON)
	r := newTestReader(t, inst, testVideoFlowID)
	if err := r.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestReaderHandleNilAfterClose(t *testing.T) {
	inst := newTestInstance(t)
	_ = newTestWriter(t, inst, testVideoFlowJSON)
	r := newTestReader(t, inst, testVideoFlowID)
	if r.Handle() == nil {
		t.Fatal("Handle() returned nil for open reader")
	}
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if r.Handle() != nil {
		t.Fatal("Handle() returned non-nil after Close")
	}
}

func TestReaderInfoMatchesWriter(t *testing.T) {
	inst := newTestInstance(t)
	w := newTestWriter(t, inst, testVideoFlowJSON)
	r := newTestReader(t, inst, testVideoFlowID)

	info, err := r.Info()
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if got := info.Config.Common.Format; got != w.Config().Common.Format {
		t.Fatalf("reader Format = %s, want %s", got, w.Config().Common.Format)
	}
	if got := info.Config.Common.GrainRate; got != w.Config().Common.GrainRate {
		t.Fatalf("reader GrainRate = %+v, want %+v", got, w.Config().Common.GrainRate)
	}

	cfg, err := r.Config()
	if err != nil {
		t.Fatalf("Config: %v", err)
	}
	if cfg.Common.Format != info.Config.Common.Format {
		t.Fatalf("Config().Format = %s, Info().Format = %s", cfg.Common.Format, info.Config.Common.Format)
	}
	if _, err := r.Runtime(); err != nil {
		t.Fatalf("Runtime: %v", err)
	}
}

func TestReaderGetMaxReadLengthSamplesAudio(t *testing.T) {
	inst := newTestInstance(t)
	_ = newTestWriter(t, inst, testAudioFlowJSON)
	r := newTestReader(t, inst, "b3bb5be7-9fe9-4324-a5bb-4c70e1084449")
	n, err := r.GetMaxReadLengthSamples()
	if err != nil {
		t.Fatalf("GetMaxReadLengthSamples: %v", err)
	}
	if n == 0 {
		t.Fatal("GetMaxReadLengthSamples returned 0")
	}
}

func TestReaderMethodsAfterClose(t *testing.T) {
	inst := newTestInstance(t)
	_ = newTestWriter(t, inst, testVideoFlowJSON)
	r := newTestReader(t, inst, testVideoFlowID)
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := r.Info(); !errors.Is(err, ErrClosed) {
		t.Errorf("Info after Close: %v, want ErrClosed", err)
	}
	if _, err := r.Config(); !errors.Is(err, ErrClosed) {
		t.Errorf("Config after Close: %v, want ErrClosed", err)
	}
	if _, err := r.Runtime(); !errors.Is(err, ErrClosed) {
		t.Errorf("Runtime after Close: %v, want ErrClosed", err)
	}
	if _, err := r.GetGrain(0, time.Millisecond); !errors.Is(err, ErrClosed) {
		t.Errorf("GetGrain after Close: %v, want ErrClosed", err)
	}
	if _, err := r.GetGrainSlice(0, 0, time.Millisecond); !errors.Is(err, ErrClosed) {
		t.Errorf("GetGrainSlice after Close: %v, want ErrClosed", err)
	}
	if _, err := r.GetGrainNonBlocking(0); !errors.Is(err, ErrClosed) {
		t.Errorf("GetGrainNonBlocking after Close: %v, want ErrClosed", err)
	}
	if _, err := r.GetGrainSliceNonBlocking(0, 0); !errors.Is(err, ErrClosed) {
		t.Errorf("GetGrainSliceNonBlocking after Close: %v, want ErrClosed", err)
	}
	if _, err := r.GetMaxReadLengthSamples(); !errors.Is(err, ErrClosed) {
		t.Errorf("GetMaxReadLengthSamples after Close: %v, want ErrClosed", err)
	}
	if _, err := r.GetSamples(0, 1, time.Millisecond); !errors.Is(err, ErrClosed) {
		t.Errorf("GetSamples after Close: %v, want ErrClosed", err)
	}
	if _, err := r.GetSamplesNonBlocking(0, 1); !errors.Is(err, ErrClosed) {
		t.Errorf("GetSamplesNonBlocking after Close: %v, want ErrClosed", err)
	}
}

func TestReaderGetSamplesInvalidCount(t *testing.T) {
	inst := newTestInstance(t)
	_ = newTestWriter(t, inst, testAudioFlowJSON)
	r := newTestReader(t, inst, "b3bb5be7-9fe9-4324-a5bb-4c70e1084449")
	if _, err := r.GetSamples(0, 0, time.Millisecond); !errors.Is(err, ErrInvalidArg) {
		t.Errorf("GetSamples(_, 0): %v, want ErrInvalidArg", err)
	}
	if _, err := r.GetSamplesNonBlocking(0, -1); !errors.Is(err, ErrInvalidArg) {
		t.Errorf("GetSamplesNonBlocking(_, -1): %v, want ErrInvalidArg", err)
	}
}
