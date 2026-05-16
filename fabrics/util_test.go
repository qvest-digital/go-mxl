package fabrics

import (
	"testing"
	"time"
)

func TestTimeoutMsClamps(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want uint16
	}{
		{0, 0},
		{-time.Second, 0},
		{500 * time.Microsecond, 1},
		{1 * time.Millisecond, 1},
		{500 * time.Millisecond, 500},
		{2 * time.Minute, 0xFFFF},
	}
	for _, c := range cases {
		if got := timeoutMs(c.in); got != c.want {
			t.Errorf("timeoutMs(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}
