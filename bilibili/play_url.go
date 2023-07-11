package bilibili

import (
	"fmt"
	"github.com/keuin/slbr/types"
)

func (b Bilibili) GetStreamingInfo(roomId types.RoomId) (resp types.RoomUrlInfoResponse, err error) {
	url := fmt.Sprintf("https://api.live.bilibili.com/room/v1/Room/playUrl?"+
		"cid=%d&otype=json&qn=10000&platform=web", roomId)
	return callGet[types.RoomUrlInfoResponse](b, url)
}
