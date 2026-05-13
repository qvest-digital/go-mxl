// Package mxl provides Go bindings to the MXL (Media eXchange Layer) C SDK.
//
// MXL is a shared-memory publish/subscribe layer for exchanging uncompressed
// video, audio, and ancillary data between co-located media functions. Flows
// live as tmpfs-backed ring buffers; readers and writers communicate via
// memory mapping and futex wakeups, with no copies in the data path.
//
// The binding covers the full libmxl public API:
//
//   - Instance management (NewInstance, IsTmpFs, GarbageCollect, FlowDef, ...)
//   - Discrete-flow reads:  Reader.GetGrain[Slice]([NonBlocking])
//   - Discrete-flow writes: Writer.OpenGrain / Commit / Cancel
//   - Continuous-flow reads:  Reader.GetSamples[NonBlocking]
//   - Continuous-flow writes: Writer.OpenSamples / Commit / Cancel
//   - Synchronization groups: Instance.NewSyncGroup
//   - Time/index helpers: Now, CurrentIndex, IndexToTimestamp, ...
//
// # Build requirements
//
//   - libmxl must be installed and visible to pkg-config. Run
//     `pkg-config --cflags --libs libmxl` to verify before `go build`;
//     it should return both -I/-L flags and `-lmxl -lmxl-common` plus
//     the dependent spdlog/fmt libraries.
//   - cgo must be enabled (CGO_ENABLED=1, default on native builds).
//
// # Typical usage
//
//	inst, err := mxl.NewInstance("/dev/shm/mxl", "")
//	if err != nil { return err }
//	defer inst.Close()
//
//	r, err := inst.NewReader("<flow-uuid>")
//	if err != nil { return err }
//	defer r.Close()
//
//	info, _ := r.Info()
//	for {
//	    idx := mxl.CurrentIndex(info.Config.Common.GrainRate)
//	    g, err := r.GetGrain(idx, 200*time.Millisecond)
//	    if errors.Is(err, mxl.ErrTimeout) { continue }
//	    if err != nil { return err }
//	    handle(g.Payload) // borrowed; do not retain past next read
//	}
//
// # Borrowed payloads
//
// Byte slices returned by reads (Grain.Payload, SamplesView fragments) and
// by writer-side OpenGrain/OpenSamples aliase libmxl's shared memory
// directly. They are only valid until the next read, the matching
// Commit/Cancel, or the Reader/Writer being closed. Use Grain.Copy() or
// SamplesView.CopyChannel() to retain data.
package mxl
