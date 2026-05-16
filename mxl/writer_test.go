package mxl

import (
	"errors"
	"testing"
)

// testVideoFlowJSON is a minimal but libmxl-acceptable v210/1080p29 flow
// definition, reused across mxl unit tests. The id is fixed because every
// test runs against a freshly created tmpfs domain anyway.
const testVideoFlowJSON = `{
  "description": "go-mxl unit test, 1080p29",
  "id": "5fbec3b1-1b0f-417d-9059-8b94a47197ed",
  "format": "urn:x-nmos:format:video",
  "label": "go-mxl unit test video",
  "tags": { "urn:x-nmos:tag:grouphint/v1.0": ["go-mxl unit test:Video"] },
  "parents": [],
  "media_type": "video/v210",
  "grain_rate": { "numerator": 30000, "denominator": 1001 },
  "frame_width": 1920,
  "frame_height": 1080,
  "interlace_mode": "progressive",
  "colorspace": "BT709",
  "components": [
    { "name": "Y",  "width": 1920, "height": 1080, "bit_depth": 10 },
    { "name": "Cb", "width": 960,  "height": 1080, "bit_depth": 10 },
    { "name": "Cr", "width": 960,  "height": 1080, "bit_depth": 10 }
  ]
}`

const testVideoFlowID = "5fbec3b1-1b0f-417d-9059-8b94a47197ed"

const testAudioFlowJSON = `{
  "description": "go-mxl unit test audio",
  "format": "urn:x-nmos:format:audio",
  "label": "go-mxl unit test audio",
  "id": "b3bb5be7-9fe9-4324-a5bb-4c70e1084449",
  "tags": { "urn:x-nmos:tag:grouphint/v1.0": ["go-mxl unit test:Audio"] },
  "media_type": "audio/float32",
  "sample_rate": { "numerator": 48000 },
  "channel_count": 2,
  "bit_depth": 32,
  "parents": []
}`

func newTestWriter(t *testing.T, inst *Instance, def string) *Writer {
	t.Helper()
	w, _, err := inst.NewWriter(def)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	t.Cleanup(func() { w.Close() })
	return w
}

func TestNewWriterBadJSON(t *testing.T) {
	inst := newTestInstance(t)
	if _, _, err := inst.NewWriter("not json"); err == nil {
		t.Fatal("NewWriter(\"not json\") returned nil error")
	}
}

func TestNewWriterMissingRequiredField(t *testing.T) {
	inst := newTestInstance(t)
	// Tags are required by libmxl's FlowParser; drop them.
	def := `{"id":"00000000-0000-0000-0000-000000000000","format":"urn:x-nmos:format:video"}`
	if _, _, err := inst.NewWriter(def); err == nil {
		t.Fatal("NewWriter(missing tags) returned nil error")
	}
}

func TestWriterCloseIdempotent(t *testing.T) {
	inst := newTestInstance(t)
	w := newTestWriter(t, inst, testVideoFlowJSON)
	if err := w.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestWriterReportsFlowFormat(t *testing.T) {
	inst := newTestInstance(t)
	wv := newTestWriter(t, inst, testVideoFlowJSON)
	if !wv.Config().Common.Format.IsDiscrete() {
		t.Fatalf("video flow Format = %s, expected discrete", wv.Config().Common.Format)
	}

	wa := newTestWriter(t, inst, testAudioFlowJSON)
	if wa.Config().Common.Format != FormatAudio {
		t.Fatalf("audio flow Format = %s, expected audio", wa.Config().Common.Format)
	}
}

func TestWriterHandleNilAfterClose(t *testing.T) {
	inst := newTestInstance(t)
	w := newTestWriter(t, inst, testVideoFlowJSON)
	if w.Handle() == nil {
		t.Fatal("Handle() returned nil for open writer")
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if w.Handle() != nil {
		t.Fatal("Handle() returned non-nil after Close")
	}
}

func TestWriterMethodsAfterClose(t *testing.T) {
	inst := newTestInstance(t)
	w := newTestWriter(t, inst, testVideoFlowJSON)
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := w.GrainInfo(0); !errors.Is(err, ErrClosed) {
		t.Errorf("GrainInfo after Close: %v, want ErrClosed", err)
	}
	if _, err := w.OpenGrain(0); !errors.Is(err, ErrClosed) {
		t.Errorf("OpenGrain after Close: %v, want ErrClosed", err)
	}
	if _, err := w.GetMaxWriteLengthSamples(); !errors.Is(err, ErrClosed) {
		t.Errorf("GetMaxWriteLengthSamples after Close: %v, want ErrClosed", err)
	}
	if _, err := w.OpenSamples(0, 1); !errors.Is(err, ErrClosed) {
		t.Errorf("OpenSamples after Close: %v, want ErrClosed", err)
	}
}

func TestWriterOpenSamplesInvalidCount(t *testing.T) {
	inst := newTestInstance(t)
	wa := newTestWriter(t, inst, testAudioFlowJSON)
	if _, err := wa.OpenSamples(0, 0); !errors.Is(err, ErrInvalidArg) {
		t.Errorf("OpenSamples(_, 0): %v, want ErrInvalidArg", err)
	}
	if _, err := wa.OpenSamples(0, -1); !errors.Is(err, ErrInvalidArg) {
		t.Errorf("OpenSamples(_, -1): %v, want ErrInvalidArg", err)
	}
}

func TestGrainWriteAccessDoubleCommit(t *testing.T) {
	inst := newTestInstance(t)
	w := newTestWriter(t, inst, testVideoFlowJSON)
	idx := CurrentIndex(w.Config().Common.GrainRate)
	ga, err := w.OpenGrain(idx)
	if err != nil {
		t.Fatalf("OpenGrain: %v", err)
	}
	if err := ga.Commit(ga.TotalSlices, 0); err != nil {
		t.Fatalf("first Commit: %v", err)
	}
	if err := ga.Commit(ga.TotalSlices, 0); err == nil {
		t.Fatal("second Commit returned nil error")
	}
	if ga.Payload != nil {
		t.Fatal("Payload not cleared after Commit")
	}
}

func TestGrainWriteAccessCancelIdempotent(t *testing.T) {
	inst := newTestInstance(t)
	w := newTestWriter(t, inst, testVideoFlowJSON)
	idx := CurrentIndex(w.Config().Common.GrainRate)
	ga, err := w.OpenGrain(idx)
	if err != nil {
		t.Fatalf("OpenGrain: %v", err)
	}
	if err := ga.Cancel(); err != nil {
		t.Fatalf("first Cancel: %v", err)
	}
	if err := ga.Cancel(); err != nil {
		t.Fatalf("second Cancel: %v", err)
	}
	if err := ga.Commit(ga.TotalSlices, 0); err == nil {
		t.Fatal("Commit after Cancel returned nil error")
	}
}

func TestSamplesWriteAccessChannelOutOfRange(t *testing.T) {
	inst := newTestInstance(t)
	wa := newTestWriter(t, inst, testAudioFlowJSON)
	sa, err := wa.OpenSamples(CurrentIndex(wa.Config().Common.GrainRate), 480)
	if err != nil {
		t.Fatalf("OpenSamples: %v", err)
	}
	t.Cleanup(func() { sa.Cancel() })
	if _, _, err := sa.ChannelFragments(sa.ChannelCount); !errors.Is(err, ErrInvalidArg) {
		t.Errorf("ChannelFragments(out-of-range): %v, want ErrInvalidArg", err)
	}
}

func TestSamplesWriteAccessAfterCommit(t *testing.T) {
	inst := newTestInstance(t)
	wa := newTestWriter(t, inst, testAudioFlowJSON)
	sa, err := wa.OpenSamples(CurrentIndex(wa.Config().Common.GrainRate), 480)
	if err != nil {
		t.Fatalf("OpenSamples: %v", err)
	}
	if err := sa.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if err := sa.Commit(); err == nil {
		t.Fatal("second Commit returned nil error")
	}
	if _, _, err := sa.ChannelFragments(0); err == nil {
		t.Fatal("ChannelFragments after Commit returned nil error")
	}
}
