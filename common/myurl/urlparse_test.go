package myurl

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
	tests[0].Actual, err = Url("http://www.example.com/index.html").FileExtension()
	if err != nil {
		t.Fatalf("GetFileExtensionFromUrl: %v", err)
	}
	tests[1].Actual, err = Url("https://www.example.com/index.htm").FileExtension()
	if err != nil {
		t.Fatalf("GetFileExtensionFromUrl: %v", err)
	}
	tests[2].Actual, err = Url("https://www.example.com/video.flv?a=1&b=2flv").FileExtension()
	if err != nil {
		t.Fatalf("GetFileExtensionFromUrl: %v", err)
	}
	for i, tc := range tests {
		if tc.Expected != tc.Actual {
			t.Fatalf("Test %v failed: %v", i, tc)
		}
	}
}
