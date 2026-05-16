package mxl

import (
	"math"
	"testing"
	"time"
)

func TestRationalFloat64(t *testing.T) {
	cases := []struct {
		in   Rational
		want float64
	}{
		{Rational{Num: 30000, Den: 1001}, 30000.0 / 1001.0},
		{Rational{Num: 50, Den: 1}, 50.0},
		{Rational{Num: 0, Den: 1}, 0.0},
		{Rational{Num: 1, Den: 0}, 0.0},
	}
	for _, c := range cases {
		got := c.in.Float64()
		if math.Abs(got-c.want) > 1e-9 {
			t.Fatalf("%v.Float64() = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestCurrentIndexInvalidRational(t *testing.T) {
	if got := CurrentIndex(Rational{Num: 0, Den: 0}); got != UndefinedIndex {
		t.Fatalf("CurrentIndex({0,0}) = %d, want UndefinedIndex", got)
	}
}

func TestTimestampIndexRoundTrip(t *testing.T) {
	rate := Rational{Num: 30000, Den: 1001}
	const idx uint64 = 1_000_000
	ts := IndexToTimestamp(rate, idx)
	if ts == 0 {
		t.Fatalf("IndexToTimestamp returned 0 for idx=%d", idx)
	}
	if got := TimestampToIndex(rate, ts); got != idx {
		t.Fatalf("round-trip: idx %d -> ts %d -> idx %d", idx, ts, got)
	}
}

func TestNowMonotonicEnough(t *testing.T) {
	a := Now()
	time.Sleep(2 * time.Millisecond)
	b := Now()
	if b <= a {
		t.Fatalf("Now() did not advance: a=%d b=%d", a, b)
	}
}

func TestSleepForZeroOrNegativeReturns(t *testing.T) {
	start := time.Now()
	SleepFor(0)
	SleepFor(-time.Second)
	if time.Since(start) > 50*time.Millisecond {
		t.Fatalf("SleepFor(<=0) blocked")
	}
}
