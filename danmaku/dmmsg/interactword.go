package dmmsg

type InteractWordMessage struct {
	Contribution struct {
		Grade int `json:"grade"`
	} `json:"contribution"`
	DanMuScore int `json:"dmscore"`
	FansMedal  struct {
		AnchorRoomid int    `json:"anchor_roomid"`
		GuardLevel   int    `json:"guard_level"`
		IconID       int    `json:"icon_id"`
		IsLighted    int    `json:"is_lighted"`
		Color        int    `json:"medal_color"`
		ColorBorder  int    `json:"medal_color_border"`
		ColorEnd     int    `json:"medal_color_end"`
		ColorStart   int    `json:"medal_color_start"`
		Level        int    `json:"medal_level"`
		Name         string `json:"medal_name"`
		Score        int    `json:"score"`
		Special      string `json:"special"`
		TargetID     int    `json:"target_id"`
	} `json:"fans_medal"`
	Identities    []int  `json:"identities"`
	IsSpread      int    `json:"is_spread"`
	MsgType       int    `json:"msg_type"`
	PrivilegeType int    `json:"privilege_type"`
	RoomId        int    `json:"roomid"`
	Score         int64  `json:"score"`
	SpreadDesc    string `json:"spread_desc"`
	SpreadInfo    string `json:"spread_info"`
	TailIcon      int    `json:"tail_icon"`
	Timestamp     int    `json:"timestamp"`
	TriggerTime   int64  `json:"trigger_time"`
	UID           int    `json:"uid"`
	UserName      string `json:"uname"`
	UserNameColor string `json:"uname_color"`
}

type RawInteractWordMessage = BaseRawMessage[interface{}, InteractWordMessage]
