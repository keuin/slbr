/*
This file implements the auth exchange.
When Bilibili live client established the WebSocket connection successfully,
it sends this message at first. The server then responses a OpConnectOk exchange with body `{"code":0}` which indicates success.
*/
package dmpkg

import (
	"encoding/json"
	"fmt"
	"github.com/keuin/slbr/common"
)

type authInfo struct {
	UID      uint64 `json:"uid"`
	RoomId   uint64 `json:"roomid"`
	ProtoVer int    `json:"protover"`
	Platform string `json:"platform"`
	Type     int    `json:"type"`
	Key      string `json:"key"`
}

// NewAuth creates a new authentication exchange.
func NewAuth(protocol ProtocolVer, roomId common.RoomId, authKey string) (exc DanmakuExchange) {
	exc, _ = NewPlainExchange(OpConnect, authInfo{
		UID:      kUidGuest,
		RoomId:   uint64(roomId),
		ProtoVer: int(protocol),
		Platform: kPlatformWeb,
		Type:     kAuthTypeDefault,
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
