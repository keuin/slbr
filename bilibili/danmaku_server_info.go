package bilibili

import (
	"fmt"
)

type DanmakuServerInfoResponse = BaseResponse[danmakuInfo]

type danmakuInfo struct {
	Group            string  `json:"group"`
	BusinessID       int     `json:"business_id"`
	RefreshRowFactor float64 `json:"refresh_row_factor"`
	RefreshRate      int     `json:"refresh_rate"`
	MaxDelay         int     `json:"max_delay"`
	Token            string  `json:"token"`
	HostList         []struct {
		Host    string `json:"host"`
		Port    int    `json:"port"`
		WssPort int    `json:"wss_port"`
		WsPort  int    `json:"ws_port"`
	} `json:"host_list"`
}

func (b Bilibili) GetDanmakuServerInfo(roomId RoomId) (resp DanmakuServerInfoResponse, err error) {
	url := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?id=%d&type=0", roomId)
	return callGet[DanmakuServerInfoResponse](b, url)
}
