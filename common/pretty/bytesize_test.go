package pretty

import (
	"testing"
)

func TestBytes(t *testing.T) {
	tests := []struct {
		Expected string
		Actual   string
	}{
		{"128 Byte", Bytes(128)},
		{"128.00 KiB", Bytes(128 * 1024)},
		{"128.00 MiB", Bytes(128 * 1024 * 1024)},
		{"128.00 GiB", Bytes(128 * 1024 * 1024 * 1024)},
		{"128.00 TiB", Bytes(128 * 1024 * 1024 * 1024 * 1024)},
		{"131072.00 TiB", Bytes(128 * 1024 * 1024 * 1024 * 1024 * 1024)},
	}
	for i, tc := range tests {
		if tc.Expected != tc.Actual {
			t.Fatalf("Test %v failed: %v", i, tc)
		}
	}
}
