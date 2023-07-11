/*
Get live room basic status.
This is used to check initially if it is streaming or not.
*/
package bilibili

import (
	"fmt"
	"github.com/keuin/slbr/types"
)

func (b Bilibili) GetRoomPlayInfo(roomId types.RoomId) (resp types.RoomPlayInfoResponse, err error) {
	url := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo"+
		"?room_id=%d&protocol=0,1&format=0,1,2&codec=0,1&qn=0&platform=web&ptype=8&dolby=5&panorama=1", roomId)
	return callGet[types.RoomPlayInfoResponse](b, url)
}
