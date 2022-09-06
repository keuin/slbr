/*
Get live room basic status.
This is used to check initially if it is streaming or not.
*/
package bilibili

import (
	"bilibili-livestream-archiver/common"
	"fmt"
)

type LiveStatus int

const (
	Inactive  LiveStatus = 0
	Streaming LiveStatus = 1
	Playback  LiveStatus = 2
)

var liveStatusStringMap = map[LiveStatus]string{
	Inactive:  "inactive",
	Streaming: "streaming",
	Playback:  "inactive (playback)",
}

type roomPlayInfo struct {
	RoomID          uint64        `json:"room_id"`
	ShortID         uint          `json:"short_id"`
	UID             uint          `json:"uid"`
	IsHidden        bool          `json:"is_hidden"`
	IsLocked        bool          `json:"is_locked"`
	IsPortrait      bool          `json:"is_portrait"`
	LiveStatus      LiveStatus    `json:"live_status"` // 0: inactive 1: streaming 2: playback
	HiddenTill      int           `json:"hidden_till"`
	LockTill        int           `json:"lock_till"`
	Encrypted       bool          `json:"encrypted"`
	PwdVerified     bool          `json:"pwd_verified"`
	LiveTime        int           `json:"live_time"`
	RoomShield      int           `json:"room_shield"`
	AllSpecialTypes []interface{} `json:"all_special_types"`
	PlayurlInfo     interface{}   `json:"playurl_info"`
}

type RoomPlayInfoResponse = BaseResponse[roomPlayInfo]

func (s LiveStatus) IsStreaming() bool {
	return s == Streaming
}

func (s LiveStatus) String() string {
	return liveStatusStringMap[s]
}

func (b Bilibili) GetRoomPlayInfo(roomId common.RoomId) (resp RoomPlayInfoResponse, err error) {
	url := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo"+
		"?room_id=%d&protocol=0,1&format=0,1,2&codec=0,1&qn=0&platform=web&ptype=8&dolby=5&panorama=1", roomId)
	return callGet[RoomPlayInfoResponse](b, url)
}
