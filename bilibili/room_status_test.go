package bilibili

import (
	testing2 "github.com/keuin/slbr/common/testing"
	"github.com/keuin/slbr/logging"
	"log"
	"testing"
)

func TestBilibili_GetRoomPlayInfo(t *testing.T) {
	// get an online live room for testing
	liveList, err := testing2.GetLiveListForGuestUser()
	if err != nil {
		t.Fatalf("cannot get live list for testing: %v", err)
	}
	lives := liveList.Data.Data
	if len(lives) <= 0 {
		t.Fatalf("no live for guest available")
	}
	roomId := lives[0].Roomid

	logger := log.Default()
	bi := NewBilibili(logging.NewWrappedLogger(logger, "test-logger"))
	resp, err := bi.GetRoomPlayInfo(roomId)
	if err != nil {
		t.Fatalf("GetRoomPlayInfo: %v", err)
	}
	if resp.Code != 0 ||
		resp.Message != "0" ||
		resp.Data.UID <= 0 ||
		resp.Data.RoomID != uint64(roomId) ||
		resp.Data.LiveStatus != Streaming {
		t.Fatalf("Invalid GetRoomPlayInfo response: %v", resp)
	}
}
