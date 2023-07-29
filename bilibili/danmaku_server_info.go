package bilibili

import (
	"errors"
	"fmt"
	"github.com/keuin/slbr/types"
	"net/url"
)

const apiUrlPrefix = "https://api.live.bilibili.com"

func (b *Bilibili) GetDanmakuServerInfo(roomId types.RoomId) (resp types.DanmakuServerInfoResponse, err error) {
	u := fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getDanmuInfo?id=%d&type=0", roomId)
	return callGet[types.DanmakuServerInfoResponse](b, u)
}

// GetBUVID initializes cookie `buvid3`. If success, returns its value.
func (b *Bilibili) GetBUVID() (string, error) {
	const u = "https://data.bilibili.com/v/web/web_page_view"
	_, _, err := callGetRaw(b, u)
	if err != nil {
		return "", err
	}
	uu, _ := url.Parse(apiUrlPrefix)
	cookies := b.http.Jar.Cookies(uu)
	var buvid3 *string
	for _, c := range cookies {
		if c.Name == "buvid3" {
			buvid3 = &c.Value
		}
	}
	if buvid3 == nil {
		return "", errors.New("failed to get buvid3")
	}
	return *buvid3, nil
}

// GetLiveBUVID initializes cookie `LIVE_BUVID`. This should be called before GetDanmakuServerInfo.
func (b *Bilibili) GetLiveBUVID(roomId types.RoomId) (resp types.WebBannerResponse, err error) {
	u := fmt.Sprintf("https://api.live.bilibili.com/activity/v1/Common/webBanner?"+
		"platform=web&position=6&roomid=%d&area_v2_parent_id=0&area_v2_id=0&from=", roomId)
	resp, err = callGet[types.WebBannerResponse](b, u)
	if err == nil {
		uu, _ := url.Parse(apiUrlPrefix)
		b.logger.Info("Cookie info: %v", b.http.Jar.Cookies(uu))
	}
	return resp, err
}
