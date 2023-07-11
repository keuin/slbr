package types

type LiveList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Count int `json:"count"`
		Data  []struct {
			Face     string `json:"face"`
			Link     string `json:"link"`
			Roomid   RoomId `json:"roomid"`
			Roomname string `json:"roomname"`
			Nickname string `json:"nickname"`
		} `json:"data"`
	} `json:"data"`
}
