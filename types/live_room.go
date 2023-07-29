package types

type RoomUrlInfoResponse = BaseResponse[roomUrlInfo]

type WebBannerResponse = BaseResponse[interface{}]

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

type roomProfile struct {
	UID              int      `json:"uid"`
	RoomID           RoomId   `json:"room_id"`
	ShortID          int      `json:"short_id"`
	Attention        int      `json:"attention"`
	Online           int      `json:"online"`
	IsPortrait       bool     `json:"is_portrait"`
	Description      string   `json:"description"`
	LiveStatus       int      `json:"live_status"`
	AreaID           int      `json:"area_id"`
	ParentAreaID     int      `json:"parent_area_id"`
	ParentAreaName   string   `json:"parent_area_name"`
	OldAreaID        int      `json:"old_area_id"`
	Background       string   `json:"background"`
	Title            string   `json:"title"`
	UserCover        string   `json:"user_cover"`
	Keyframe         string   `json:"keyframe"`
	IsStrictRoom     bool     `json:"is_strict_room"`
	LiveTime         string   `json:"live_time"`
	Tags             string   `json:"tags"`
	IsAnchor         int      `json:"is_anchor"`
	RoomSilentType   string   `json:"room_silent_type"`
	RoomSilentLevel  int      `json:"room_silent_level"`
	RoomSilentSecond int      `json:"room_silent_second"`
	AreaName         string   `json:"area_name"`
	Pendants         string   `json:"pendants"`
	AreaPendants     string   `json:"area_pendants"`
	HotWords         []string `json:"hot_words"`
	HotWordsStatus   int      `json:"hot_words_status"`
	Verify           string   `json:"verify"`
	NewPendants      struct {
		Frame struct {
			Name       string `json:"name"`
			Value      string `json:"value"`
			Position   int    `json:"position"`
			Desc       string `json:"desc"`
			Area       int    `json:"area"`
			AreaOld    int    `json:"area_old"`
			BgColor    string `json:"bg_color"`
			BgPic      string `json:"bg_pic"`
			UseOldArea bool   `json:"use_old_area"`
		} `json:"frame"`
		Badge struct {
			Name     string `json:"name"`
			Position int    `json:"position"`
			Value    string `json:"value"`
			Desc     string `json:"desc"`
		} `json:"badge"`
		MobileFrame struct {
			Name       string `json:"name"`
			Value      string `json:"value"`
			Position   int    `json:"position"`
			Desc       string `json:"desc"`
			Area       int    `json:"area"`
			AreaOld    int    `json:"area_old"`
			BgColor    string `json:"bg_color"`
			BgPic      string `json:"bg_pic"`
			UseOldArea bool   `json:"use_old_area"`
		} `json:"mobile_frame"`
		MobileBadge interface{} `json:"mobile_badge"`
	} `json:"new_pendants"`
	UpSession            string `json:"up_session"`
	PkStatus             int    `json:"pk_status"`
	PkID                 int    `json:"pk_id"`
	BattleID             int    `json:"battle_id"`
	AllowChangeAreaTime  int    `json:"allow_change_area_time"`
	AllowUploadCoverTime int    `json:"allow_upload_cover_time"`
	StudioInfo           struct {
		Status     int           `json:"status"`
		MasterList []interface{} `json:"master_list"`
	} `json:"studio_info"`
}

type RoomProfileResponse = BaseResponse[roomProfile]

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
