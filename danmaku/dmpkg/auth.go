/*
This file implements the auth exchange.
When Bilibili live client established the WebSocket connection successfully,
it sends this message at first. The server then responses a OpConnectOk exchange with body `{"code":0}` which indicates success.
*/
package dmpkg

import (
	"encoding/json"
	"fmt"
	"github.com/keuin/slbr/types"
)

type authInfo struct {
	UID      uint64       `json:"uid"`
	RoomId   types.RoomId `json:"roomid"`
	ProtoVer int          `json:"protover"`
	BUVID3   string       `json:"buvid"`
	Platform string       `json:"platform"`
	Type     int          `json:"type"`
	Key      string       `json:"key"`
}

// NewAuth creates a new authentication exchange.
func NewAuth(protocol ProtocolVer, roomId types.RoomId, authKey, buvid3 string) (exc DanmakuExchange) {
	exc, _ = NewPlainExchange(OpConnect, authInfo{
		UID:      UidGuest,
		RoomId:   roomId,
		ProtoVer: int(protocol),
		BUVID3:   buvid3,
		Platform: PlatformWeb,
		Type:     AuthTypeDefault,
		Key:      authKey,
	})
	return
}

func IsAuthOk(serverResponse DanmakuExchange) (bool, error) {
	if op := serverResponse.Operation; op != OpConnectOk {
		return false, fmt.Errorf("server operation is not OpConnectOk: %v", op)
	}
	var body struct {
		Code int `json:"code"`
	}
	body.Code = 1
	err := json.Unmarshal(serverResponse.Body, &body)
	if err != nil {
		return false, fmt.Errorf("JSON decode error: %w", err)
	}
	if c := body.Code; c != 0 {
		return false, fmt.Errorf("server response code is non-zero: %v", c)
	}
	return true, nil
}
