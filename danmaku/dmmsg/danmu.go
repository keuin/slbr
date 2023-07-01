package dmmsg

/*
Decoder of raw danmaku messages.
*/

import (
	"fmt"
)

type RawDanMuMessage = BaseRawMessage[[]interface{}, interface{}]

type DanMuMessage struct {
	Content    string
	SourceUser struct {
		Nickname string
		UID      int64
	}
}

func (dm DanMuMessage) String() string {
	return fmt.Sprintf("(user: %v, uid: %v) %v",
		dm.SourceUser.Nickname, dm.SourceUser.UID, dm.Content)
}

const InvalidDanmakuJson = "invalid danmaku JSON document"

func ParseDanmakuMessage(body RawDanMuMessage) (dmm DanMuMessage, err error) {
	if len(body.Info) != 16 {
		err = fmt.Errorf("%s: \"info\" length != 16", InvalidDanmakuJson)
		return
	}

	dmm.Content, err = castValue[string](body.Info[1])
	if err != nil {
		return
	}

	userInfo, err := castValue[[]interface{}](body.Info[2])

	var ok bool
	uid, ok := userInfo[0].(float64)
	if !ok {
		err = fmt.Errorf("%s: uid is not a float64: %v", InvalidDanmakuJson, userInfo[0])
		return
	}
	dmm.SourceUser.UID = int64(uid)

	dmm.SourceUser.Nickname, ok = userInfo[1].(string)
	if !ok {
		err = fmt.Errorf("%s: nickname is not a string: %v", InvalidDanmakuJson, userInfo[1])
		return
	}
	return
}
