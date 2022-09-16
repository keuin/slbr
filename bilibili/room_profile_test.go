package bilibili

import (
	"github.com/keuin/slbr/common"
	"github.com/keuin/slbr/logging"
	"log"
	"testing"
)

func TestBilibili_GetRoomProfile(t *testing.T) {
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
	resp, err := bi.GetRoomProfile(roomId)
	if err != nil {
		t.Fatalf("GetRoomProfile: %v", err)
	}
	if resp.Code != 0 ||
		resp.Message != "ok" ||
		resp.Data.UID <= 0 ||
		resp.Data.RoomID != int(roomId) ||
		resp.Data.LiveStatus != int(Streaming) ||
		resp.Data.Title == "" {
		t.Fatalf("Invalid GetRoomProfile response: %v", resp)
	}
}
