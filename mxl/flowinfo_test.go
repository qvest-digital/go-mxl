package mxl

import "testing"

func TestDataFormatIsDiscrete(t *testing.T) {
	cases := []struct {
		f    DataFormat
		want bool
	}{
		{FormatVideo, true},
		{FormatData, true},
		{FormatAudio, false},
		{FormatUnspecified, false},
	}
	for _, c := range cases {
		if got := c.f.IsDiscrete(); got != c.want {
			t.Errorf("%s.IsDiscrete() = %v, want %v", c.f, got, c.want)
		}
	}
}

func TestDataFormatString(t *testing.T) {
	cases := []struct {
		f    DataFormat
		want string
	}{
		{FormatVideo, "video"},
		{FormatAudio, "audio"},
		{FormatData, "data"},
		{FormatUnspecified, "unspecified"},
		{DataFormat(0xDEADBEEF), "unknown"},
	}
	for _, c := range cases {
		if got := c.f.String(); got != c.want {
			t.Errorf("DataFormat(%d).String() = %q, want %q", c.f, got, c.want)
		}
	}
}
