package bilibili

import (
	testing2 "github.com/keuin/slbr/common/testing"
	"github.com/keuin/slbr/logging"
	"log"
	"testing"
)

func TestBilibili_GetStreamingInfo(t *testing.T) {
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
	_, err = bi.GetBUVID()
	if err != nil {
		t.Fatalf("GetBUVID: %v", err)
	}
	info, err := bi.GetStreamingInfo(roomId)
	if err != nil {
		t.Fatalf("GetStreamingInfo: %v", err)
	}
	if info.Code != 0 ||
		info.Message != "0" ||
		len(info.Data.URLs) <= 0 ||
		len(info.Data.AcceptQuality) <= 0 ||
		len(info.Data.QualityDescription) <= 0 {
		t.Fatalf("Invalid GetStreamingInfo response: %v", info)
	}
}
