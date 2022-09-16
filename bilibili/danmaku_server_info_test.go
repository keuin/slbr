package bilibili

import (
	"github.com/keuin/slbr/common"
	"github.com/keuin/slbr/logging"
	"log"
	"testing"
)

func TestBilibili_GetDanmakuServerInfo(t *testing.T) {
	// get an online live room for testing
	liveList, err := common.GetLiveListForGuestUser()
	if err != nil {
		t.Fatalf("Cannot get live list for testing: %v", err)
	}
	lives := liveList.Data.Data
	if len(lives) <= 0 {
		t.Fatalf("No available live for guest user")
	}
	roomId := common.RoomId(lives[0].Roomid)

	logger := log.Default()
	bi := NewBilibili(logging.NewWrappedLogger(logger, "test-logger"))
	dmInfo, err := bi.GetDanmakuServerInfo(roomId)
	if err != nil {
		t.Fatalf("GetDanmakuServerInfo: %v", err)
	}
	if dmInfo.Code != 0 ||
		dmInfo.Message != "0" ||
		len(dmInfo.Data.Token) < 10 ||
		len(dmInfo.Data.HostList) <= 0 {
		t.Fatalf("Invalid GetDanmakuServerInfo response: %v", dmInfo)
	}
	for _, h := range dmInfo.Data.HostList {
		if h.Port == 0 || h.WssPort == 0 || h.WsPort == 0 || h.Host == "" {
			t.Fatalf("Invalid host: %v", h)
		}
	}
}
