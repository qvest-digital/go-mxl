package mxl

/*
#include <mxl/flowinfo.h>
#include <mxl/dataformat.h>
*/
import "C"

// DataFormat mirrors mxlDataFormat. Discrete formats (video, data) use grain
// I/O; continuous formats (audio) use sample I/O.
type DataFormat uint32

const (
	FormatUnspecified DataFormat = C.MXL_DATA_FORMAT_UNSPECIFIED
	FormatVideo       DataFormat = C.MXL_DATA_FORMAT_VIDEO
	FormatAudio       DataFormat = C.MXL_DATA_FORMAT_AUDIO
	FormatData        DataFormat = C.MXL_DATA_FORMAT_DATA
)

// IsDiscrete reports whether this format uses discrete grains (vs continuous
// samples). Video and data are discrete; audio is continuous.
func (f DataFormat) IsDiscrete() bool {
	return f == FormatVideo || f == FormatData
}

func (f DataFormat) String() string {
	switch f {
	case FormatVideo:
		return "video"
	case FormatAudio:
		return "audio"
	case FormatData:
		return "data"
	case FormatUnspecified:
		return "unspecified"
	default:
		return "unknown"
	}
}

// PayloadLocation indicates whether grain memory lives in host RAM or on a
// device (e.g. GPU).
type PayloadLocation uint32

const (
	PayloadHostMemory   PayloadLocation = C.MXL_PAYLOAD_LOCATION_HOST_MEMORY
	PayloadDeviceMemory PayloadLocation = C.MXL_PAYLOAD_LOCATION_DEVICE_MEMORY
)

// Rational is a fraction used by MXL for edit/sample rates.
type Rational struct {
	Num int64
	Den int64
}

// Float64 returns Num/Den as a float64. Returns 0 if Den is 0.
func (r Rational) Float64() float64 {
	if r.Den == 0 {
		return 0
	}
	return float64(r.Num) / float64(r.Den)
}

// CommonFlowConfig is the format-independent part of a flow's configuration.
type CommonFlowConfig struct {
	ID                    [16]byte
	Format                DataFormat
	Flags                 uint32
	GrainRate             Rational
	MaxCommitBatchSizeHint uint32
	MaxSyncBatchSizeHint   uint32
	PayloadLocation       PayloadLocation
	DeviceIndex           int32
}

// DiscreteFlowConfig holds video/data-specific config.
type DiscreteFlowConfig struct {
	SliceSizes [4]uint32 // MXL_MAX_PLANES_PER_GRAIN = 4
	GrainCount uint32
}

// ContinuousFlowConfig holds audio-specific config.
type ContinuousFlowConfig struct {
	ChannelCount uint32
	BufferLength uint32
}

// FlowConfig is the immutable description of a flow. Exactly one of
// Discrete/Continuous is meaningful, determined by Common.Format.
type FlowConfig struct {
	Common     CommonFlowConfig
	Discrete   DiscreteFlowConfig   // valid iff Common.Format.IsDiscrete()
	Continuous ContinuousFlowConfig // valid iff !Common.Format.IsDiscrete()
}

// FlowRuntime is the mutable state of a flow: where the ring head is, last
// reader/writer activity timestamps.
type FlowRuntime struct {
	HeadIndex     uint64
	LastWriteTime uint64 // ns since epoch
	LastReadTime  uint64 // ns since epoch
}

// FlowInfo is the full flow header (config + runtime).
type FlowInfo struct {
	Version uint32
	Size    uint32
	Config  FlowConfig
	Runtime FlowRuntime
}

func goFlowConfig(c *C.mxlFlowConfigInfo) FlowConfig {
	out := FlowConfig{
		Common: CommonFlowConfig{
			Format:                 DataFormat(c.common.format),
			Flags:                  uint32(c.common.flags),
			GrainRate:              Rational{Num: int64(c.common.grainRate.numerator), Den: int64(c.common.grainRate.denominator)},
			MaxCommitBatchSizeHint: uint32(c.common.maxCommitBatchSizeHint),
			MaxSyncBatchSizeHint:   uint32(c.common.maxSyncBatchSizeHint),
			PayloadLocation:        PayloadLocation(c.common.payloadLocation),
			DeviceIndex:            int32(c.common.deviceIndex),
		},
	}
	for i := 0; i < 16; i++ {
		out.Common.ID[i] = byte(c.common.id[i])
	}
	// The union: pick the right view based on format. The C union has the
	// same offset for both branches, so reading the matching branch is safe.
	if out.Common.Format.IsDiscrete() {
		d := (*C.mxlDiscreteFlowConfigInfo)(discreteFromUnion(c))
		out.Discrete.GrainCount = uint32(d.grainCount)
		for i := 0; i < 4; i++ {
			out.Discrete.SliceSizes[i] = uint32(d.sliceSizes[i])
		}
	} else if out.Common.Format == FormatAudio {
		cont := (*C.mxlContinuousFlowConfigInfo)(continuousFromUnion(c))
		out.Continuous.ChannelCount = uint32(cont.channelCount)
		out.Continuous.BufferLength = uint32(cont.bufferLength)
	}
	return out
}

func goFlowRuntime(r *C.mxlFlowRuntimeInfo) FlowRuntime {
	return FlowRuntime{
		HeadIndex:     uint64(r.headIndex),
		LastWriteTime: uint64(r.lastWriteTime),
		LastReadTime:  uint64(r.lastReadTime),
	}
}

func goFlowInfo(fi *C.mxlFlowInfo) FlowInfo {
	return FlowInfo{
		Version: uint32(fi.version),
		Size:    uint32(fi.size),
		Config:  goFlowConfig(&fi.config),
		Runtime: goFlowRuntime(&fi.runtime),
	}
}
