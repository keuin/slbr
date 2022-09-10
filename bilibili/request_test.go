package bilibili

import (
	"bilibili-livestream-archiver/logging"
	"log"
	"testing"
)

func Test_callGet(t *testing.T) {
	// an always-fail request should not panic
	bi := NewBilibili(logging.NewWrappedLogger(log.Default(), "main"))
	_, err := callGet[BaseResponse[struct{}]](bi, "https://256.256.256.256")
	if err == nil {
		t.Fatalf("the artificial request should fail, but it haven't")
	}
}
