//go:build mxl_integration

package mxl_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/qvest-digital/go-mxl/mxl"
)

const videoFlowJSON = `{
  "description": "go-mxl test, 1080p29",
  "id": "5fbec3b1-1b0f-417d-9059-8b94a47197ed",
  "format": "urn:x-nmos:format:video",
  "label": "go-mxl test video",
  "tags": { "urn:x-nmos:tag:grouphint/v1.0": ["go-mxl test:Video"] },
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

const videoFlowID = "5fbec3b1-1b0f-417d-9059-8b94a47197ed"

const audioFlowJSON = `{
  "description": "go-mxl test audio",
  "format": "urn:x-nmos:format:audio",
  "label": "go-mxl test audio",
  "id": "b3bb5be7-9fe9-4324-a5bb-4c70e1084449",
  "tags": { "urn:x-nmos:tag:grouphint/v1.0": ["go-mxl test:Audio"] },
  "media_type": "audio/float32",
  "sample_rate": { "numerator": 48000 },
  "channel_count": 2,
  "bit_depth": 32,
  "parents": []
}`

const audioFlowID = "b3bb5be7-9fe9-4324-a5bb-4c70e1084449"

func newDomain(t *testing.T) *mxl.Instance {
	t.Helper()
	if _, err := os.Stat("/dev/shm"); err != nil {
		t.Skip("/dev/shm not present")
	}
	dir, err := os.MkdirTemp("/dev/shm", "go-mxl-it-*")
	if err != nil {
		t.Skipf("cannot create dir in /dev/shm: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	inst, err := mxl.NewInstance(dir, "")
	if err != nil {
		t.Fatalf("NewInstance(%q): %v", dir, err)
	}
	t.Cleanup(func() { inst.Close() })
	return inst
}

func TestVideoGrainRoundTrip(t *testing.T) {
	inst := newDomain(t)

	w, created, err := inst.NewWriter(videoFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	t.Cleanup(func() { w.Close() })
	if !created {
		t.Fatal("expected flow to be freshly created")
	}
	cfg := w.Config()
	if !cfg.Common.Format.IsDiscrete() {
		t.Fatalf("expected discrete flow, got %s", cfg.Common.Format)
	}

	r, err := inst.NewReader(videoFlowID)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	t.Cleanup(func() { r.Close() })

	idx := mxl.CurrentIndex(cfg.Common.GrainRate)
	if idx == mxl.UndefinedIndex {
		t.Fatal("CurrentIndex returned UndefinedIndex")
	}

	ga, err := w.OpenGrain(idx)
	if err != nil {
		t.Fatalf("OpenGrain(%d): %v", idx, err)
	}
	if int(ga.GrainSize) != len(ga.Payload) {
		t.Fatalf("GrainSize=%d but len(Payload)=%d", ga.GrainSize, len(ga.Payload))
	}
	for i := range ga.Payload {
		ga.Payload[i] = byte((uint64(i) + idx) & 0xFF)
	}
	if err := ga.Commit(ga.TotalSlices, 0); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	g, err := r.GetGrain(idx, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("GetGrain: %v", err)
	}
	if g.Index != idx {
		t.Fatalf("read grain index = %d, want %d", g.Index, idx)
	}
	if !g.Complete() {
		t.Fatalf("read grain not complete: %d/%d", g.ValidSlices, g.TotalSlices)
	}
	if got, want := uint64(g.Payload[0]), idx&0xFF; got != want {
		t.Fatalf("payload[0] = %d, want %d", got, want)
	}
	cp := g.Copy()
	if len(cp) != len(g.Payload) {
		t.Fatalf("Copy length %d != payload length %d", len(cp), len(g.Payload))
	}
}

func TestWriterCancelGrainNotPublished(t *testing.T) {
	inst := newDomain(t)
	w, _, err := inst.NewWriter(videoFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	t.Cleanup(func() { w.Close() })

	idx := mxl.CurrentIndex(w.Config().Common.GrainRate)
	ga, err := w.OpenGrain(idx)
	if err != nil {
		t.Fatalf("OpenGrain: %v", err)
	}
	if err := ga.Cancel(); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if err := ga.Cancel(); err != nil {
		t.Fatalf("second Cancel: %v", err)
	}
	if err := ga.Commit(ga.TotalSlices, 0); err == nil {
		t.Fatal("Commit after Cancel returned nil error")
	}
}

func TestAudioSamplesRoundTrip(t *testing.T) {
	inst := newDomain(t)
	w, _, err := inst.NewWriter(audioFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	t.Cleanup(func() { w.Close() })

	cfg := w.Config()
	if cfg.Common.Format != mxl.FormatAudio {
		t.Fatalf("expected audio flow, got %s", cfg.Common.Format)
	}

	r, err := inst.NewReader(audioFlowID)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	t.Cleanup(func() { r.Close() })

	// ~10 ms of samples at 48 kHz.
	const batch = 480
	idx := mxl.CurrentIndex(cfg.Common.GrainRate)

	sa, err := w.OpenSamples(idx, batch)
	if err != nil {
		t.Fatalf("OpenSamples: %v", err)
	}
	if sa.ChannelCount != uint64(cfg.Continuous.ChannelCount) {
		t.Fatalf("ChannelCount %d != flow %d", sa.ChannelCount, cfg.Continuous.ChannelCount)
	}
	for ch := uint64(0); ch < sa.ChannelCount; ch++ {
		f1, f2, err := sa.ChannelFragments(ch)
		if err != nil {
			t.Fatalf("ChannelFragments(%d): %v", ch, err)
		}
		for i := range f1 {
			f1[i] = byte((i + int(ch)) & 0xFF)
		}
		base := len(f1)
		for i := range f2 {
			f2[i] = byte((base + i + int(ch)) & 0xFF)
		}
	}
	if err := sa.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	v, err := r.GetSamples(idx, batch, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("GetSamples: %v", err)
	}
	if v.ChannelCount != sa.ChannelCount {
		t.Fatalf("read ChannelCount %d != write %d", v.ChannelCount, sa.ChannelCount)
	}
	cp, err := v.CopyChannel(0)
	if err != nil {
		t.Fatalf("CopyChannel: %v", err)
	}
	if len(cp) == 0 {
		t.Fatal("CopyChannel returned empty slice")
	}
	if got, want := cp[0], byte(0); got != want {
		t.Fatalf("ch0[0] = %d, want %d", got, want)
	}
}

func TestSyncGroupGrain(t *testing.T) {
	inst := newDomain(t)
	w, _, err := inst.NewWriter(videoFlowJSON)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	t.Cleanup(func() { w.Close() })

	r, err := inst.NewReader(videoFlowID)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	t.Cleanup(func() { r.Close() })

	g, err := inst.NewSyncGroup()
	if err != nil {
		t.Fatalf("NewSyncGroup: %v", err)
	}
	t.Cleanup(func() { g.Close() })
	if err := g.AddReader(r); err != nil {
		t.Fatalf("AddReader: %v", err)
	}

	cfg := w.Config()
	idx := mxl.CurrentIndex(cfg.Common.GrainRate)
	ga, err := w.OpenGrain(idx)
	if err != nil {
		t.Fatalf("OpenGrain: %v", err)
	}
	if err := ga.Commit(ga.TotalSlices, 0); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	ts := mxl.IndexToTimestamp(cfg.Common.GrainRate, idx)
	if err := g.WaitForDataAt(ts, 200*time.Millisecond); err != nil {
		t.Fatalf("WaitForDataAt(%d): %v", ts, err)
	}

	if err := g.RemoveReader(r); err != nil {
		t.Fatalf("RemoveReader: %v", err)
	}
	// After removal an empty group should immediately satisfy any wait.
	if err := g.WaitForDataAt(ts, 50*time.Millisecond); err != nil && !errors.Is(err, mxl.ErrTimeout) {
		t.Fatalf("WaitForDataAt after RemoveReader: %v", err)
	}
}
