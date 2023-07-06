package files

import (
	"testing"
)

func TestPrettyBytes(t *testing.T) {
	tests := []struct {
		Expected string
		Actual   string
	}{
		{"128 Byte", PrettyBytes(128)},
		{"128.00 KiB", PrettyBytes(128 * 1024)},
		{"128.00 MiB", PrettyBytes(128 * 1024 * 1024)},
		{"128.00 GiB", PrettyBytes(128 * 1024 * 1024 * 1024)},
		{"128.00 TiB", PrettyBytes(128 * 1024 * 1024 * 1024 * 1024)},
		{"131072.00 TiB", PrettyBytes(128 * 1024 * 1024 * 1024 * 1024 * 1024)},
	}
	for i, tc := range tests {
		if tc.Expected != tc.Actual {
			t.Fatalf("Test %v failed: %v", i, tc)
		}
	}
}
