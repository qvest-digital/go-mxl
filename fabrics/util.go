package fabrics

import (
	"time"

	"github.com/qvest-digital/go-mxl/mxl"
)

// ErrClosed returns the parent mxl package's closed-handle sentinel.
// Wrapped in a helper so call sites stay terse and don't need to
// import mxl just for the error.
func ErrClosed() error { return mxl.ErrClosed }

// ErrInvalidArg returns the parent mxl package's invalid-argument
// sentinel.
func ErrInvalidArg() error { return mxl.ErrInvalidArg }

// timeoutMs clamps a Go duration into the uint16 millisecond shape
// libmxl-fabrics uses (0 means "non-blocking" / "no wait").
func timeoutMs(d time.Duration) uint16 {
	if d <= 0 {
		return 0
	}
	ms := d.Milliseconds()
	if ms <= 0 {
		return 1 // shortest positive wait
	}
	if ms > 0xFFFF {
		return 0xFFFF
	}
	return uint16(ms)
}
