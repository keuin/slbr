package testing

import (
	"encoding/json"
	"fmt"
	"github.com/keuin/slbr/types"
	"io"
	"net/http"
)

/*
Some utility function for test-purpose only.
*/

func GetLiveListForGuestUser() (liveList types.LiveList, err error) {
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
