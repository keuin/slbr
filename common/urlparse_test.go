package common

import "testing"

func TestGetFileExtensionFromUrl(t *testing.T) {
	tests := []struct {
		Expected string
		Actual   string
	}{
		{Expected: "html"},
		{Expected: "htm"},
		{Expected: "flv"},
	}
	var err error
	tests[0].Actual, err = GetFileExtensionFromUrl("http://www.example.com/index.html")
	if err != nil {
		t.Fatalf("GetFileExtensionFromUrl: %v", err)
	}
	tests[1].Actual, err = GetFileExtensionFromUrl("https://www.example.com/index.htm")
	if err != nil {
		t.Fatalf("GetFileExtensionFromUrl: %v", err)
	}
	tests[2].Actual, err = GetFileExtensionFromUrl("https://www.example.com/video.flv?a=1&b=2flv")
	if err != nil {
		t.Fatalf("GetFileExtensionFromUrl: %v", err)
	}
	for i, tc := range tests {
		if tc.Expected != tc.Actual {
			t.Fatalf("Test %v failed: %v", i, tc)
		}
	}
}
