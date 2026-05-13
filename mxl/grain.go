package mxl

/*
#include <mxl/flow.h>
*/
import "C"

// Grain flag bits (mirror MXL_GRAIN_FLAG_*).
const (
	// GrainFlagInvalid marks a grain whose payload should not be trusted.
	// A writer typically commits an invalid grain to advance the ring buffer
	// when it has nothing valid to deliver (e.g. an upstream input timed out).
	GrainFlagInvalid uint32 = C.MXL_GRAIN_FLAG_INVALID
)

// Slice-count sentinels for partial-grain APIs.
const (
	GrainValidSlicesAny uint16 = 0
	GrainValidSlicesAll uint16 = 0xFFFF
)

// Grain holds the metadata for one read grain plus a slice aliasing its
// payload in shared memory.
//
// IMPORTANT: Payload is valid only until the next read on the originating
// Reader. To retain data, call Copy or copy(dst, g.Payload) into a buffer
// you own.
type Grain struct {
	Index       uint64
	Flags       uint32
	GrainSize   uint32 // declared full grain size in bytes
	TotalSlices uint16
	ValidSlices uint16
	Payload     []byte // borrowed view into shared memory
}

// Complete reports whether all slices are committed (ValidSlices == TotalSlices).
func (g Grain) Complete() bool {
	return g.TotalSlices > 0 && g.ValidSlices == g.TotalSlices
}

// Invalid reports whether the writer marked this grain invalid.
func (g Grain) Invalid() bool {
	return g.Flags&GrainFlagInvalid != 0
}

// Copy returns a heap-allocated copy of Payload that is safe to retain past
// the next read.
func (g Grain) Copy() []byte {
	if len(g.Payload) == 0 {
		return nil
	}
	out := make([]byte, len(g.Payload))
	copy(out, g.Payload)
	return out
}
