package bilibili

import (
	"fmt"
	"github.com/keuin/slbr/types"
)

func (b Bilibili) GetDanmakuServerInfo(roomId types.RoomId) (resp types.DanmakuServerInfoResponse, err error) {
	url := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?id=%d&type=0", roomId)
	return callGet[types.DanmakuServerInfoResponse](b, url)
}
