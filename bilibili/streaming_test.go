package bilibili

import (
	"context"
	"errors"
	"fmt"
	"github.com/keuin/slbr/common"
	"github.com/keuin/slbr/logging"
	"log"
	"os"
	"testing"
)

func TestBilibili_CopyLiveStream(t *testing.T) {
	// get an online live room for testing
	liveList, err := common.GetLiveListForGuestUser()
	if err != nil {
		t.Fatalf("cannot get live list for testing: %v", err)
	}
	lives := liveList.Data.Data
	if len(lives) <= 0 {
		t.Fatalf("no live for guest available")
	}
	roomId := common.RoomId(lives[0].Roomid)

	logger := log.Default()
	bi := NewBilibili(logging.NewWrappedLogger(logger, "test-logger"))

	si, err := bi.GetStreamingInfo(roomId)
	if err != nil {
		t.Fatalf("GetStreamingInfo: %v", err)
	}

	// test file open failure
	testErr := fmt.Errorf("test error")
	err = bi.CopyLiveStream(context.Background(), roomId, si.Data.URLs[0], func() (*os.File, error) {
		return nil, testErr
	}, 1048576)
	if !errors.Is(err, testErr) {
		t.Fatalf("Unexpected error from CopyLiveStream: %v", err)
	}

	// TODO more tests
}
