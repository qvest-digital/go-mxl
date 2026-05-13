package mxl

/*
#include <mxl/time.h>
#include <mxl/rational.h>
*/
import "C"

import "time"

// UndefinedIndex is returned by index helpers when the input is invalid.
const UndefinedIndex uint64 = ^uint64(0)

// Now returns the current MXL clock time (TAI on Linux) in nanoseconds since
// the SMPTE ST 2059 epoch.
func Now() uint64 {
	return uint64(C.mxlGetTime())
}

// CurrentIndex returns the grain index that corresponds to "now" for a given
// edit rate. Returns UndefinedIndex if rate is invalid.
func CurrentIndex(rate Rational) uint64 {
	cr := toCRational(rate)
	return uint64(C.mxlGetCurrentIndex(&cr))
}

// NsUntilIndex returns how many nanoseconds remain until the start of the
// given grain index, given an edit rate. May be 0 if the index is in the
// past.
func NsUntilIndex(index uint64, rate Rational) uint64 {
	cr := toCRational(rate)
	return uint64(C.mxlGetNsUntilIndex(C.uint64_t(index), &cr))
}

// TimestampToIndex converts a nanoseconds-since-epoch timestamp to the
// corresponding grain index at the given edit rate.
func TimestampToIndex(rate Rational, ts uint64) uint64 {
	cr := toCRational(rate)
	return uint64(C.mxlTimestampToIndex(&cr, C.uint64_t(ts)))
}

// IndexToTimestamp converts a grain index to its ns-since-epoch timestamp.
func IndexToTimestamp(rate Rational, index uint64) uint64 {
	cr := toCRational(rate)
	return uint64(C.mxlIndexToTimestamp(&cr, C.uint64_t(index)))
}

// SleepNs sleeps for the given number of nanoseconds using libmxl's
// platform-appropriate sleep. Prefer time.Sleep in pure-Go code; this is
// here for parity with the C API.
func SleepNs(ns uint64) {
	C.mxlSleepForNs(C.uint64_t(ns))
}

// SleepUntil sleeps until the given ns-since-epoch timestamp.
func SleepUntil(ts uint64) {
	C.mxlSleepUntil(C.uint64_t(ts))
}

// SleepFor is a convenience wrapper accepting time.Duration.
func SleepFor(d time.Duration) {
	if d <= 0 {
		return
	}
	C.mxlSleepForNs(C.uint64_t(d.Nanoseconds()))
}

func toCRational(r Rational) C.mxlRational {
	return C.mxlRational{
		numerator:   C.int64_t(r.Num),
		denominator: C.int64_t(r.Den),
	}
}
