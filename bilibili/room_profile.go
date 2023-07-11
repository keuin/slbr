package bilibili

import (
	"fmt"
	"github.com/keuin/slbr/types"
)

func (b Bilibili) GetRoomProfile(roomId types.RoomId) (resp types.RoomProfileResponse, err error) {
	url := fmt.Sprintf("https://api.live.bilibili.com/room/v1/Room/get_info?room_id=%d", roomId)
	return callGet[types.RoomProfileResponse](b, url)
}
