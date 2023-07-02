package testing

import (
	"encoding/json"
	"fmt"
	"github.com/keuin/slbr/bilibili"
	"io"
	"net/http"
)

/*
Some utility function for test-purpose only.
*/

type LiveList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Count int `json:"count"`
		Data  []struct {
			Face     string          `json:"face"`
			Link     string          `json:"link"`
			Roomid   bilibili.RoomId `json:"roomid"`
			Roomname string          `json:"roomname"`
			Nickname string          `json:"nickname"`
		} `json:"data"`
	} `json:"data"`
}

func GetLiveListForGuestUser() (liveList LiveList, err error) {
	url := "https://api.live.bilibili.com/xlive/web-interface/v1/index/WebGetUnLoginRecList"
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad http response: %v", resp.StatusCode)
		return
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &liveList)
	return
}
