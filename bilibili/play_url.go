package bilibili

import (
	"fmt"
	"github.com/keuin/slbr/common"
)

type RoomUrlInfoResponse = BaseResponse[roomUrlInfo]

type roomUrlInfo struct {
	CurrentQuality       int                  `json:"current_quality"`
	AcceptQuality        []string             `json:"accept_quality"`
	CurrentQualityNumber int                  `json:"current_qn"`
	QualityDescription   []qualityDescription `json:"quality_description"`
	URLs                 []StreamingUrlInfo   `json:"durl"`
}

type qualityDescription struct {
	QualityNumber int    `json:"qn"`
	Description   string `json:"desc"`
}

type StreamingUrlInfo struct {
	URL        string `json:"url"`
	Length     int    `json:"length"`
	Order      int    `json:"order"`
	StreamType int    `json:"stream_type"`
	P2pType    int    `json:"p2p_type"`
}

func (b Bilibili) GetStreamingInfo(roomId common.RoomId) (resp RoomUrlInfoResponse, err error) {
	url := fmt.Sprintf("https://api.live.bilibili.com/room/v1/Room/playUrl?"+
		"cid=%d&otype=json&qn=10000&platform=web", roomId)
	return callGet[RoomUrlInfoResponse](b, url)
}
