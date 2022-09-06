package dmpkg

import (
	"encoding/json"
	"fmt"
	"math"
)

const kMaxBodyLength = math.MaxUint32 - kHeaderLength

// NewPlainExchange creates a new exchange with raw body specified.
// body: a struct or a raw string
func NewPlainExchange(operation Operation, body interface{}) (exc DanmakuExchange, err error) {
	var bodyData []byte

	// convert body to []byte
	if _, ok := body.(string); ok {
		// a string
		bodyData = []byte(body.(string))
	} else if _, ok := body.([]byte); ok {
		// a []byte
		copy(bodyData, body.([]byte))
	} else {
		// a JSON struct
		bodyData, err = json.Marshal(body)
		if err != nil {
			return
		}
	}

	length := uint64(kHeaderLength + len(bodyData))
	if length > kMaxBodyLength {
		err = fmt.Errorf("body is too large (> %d)", kMaxBodyLength)
		return
	}
	exc = DanmakuExchange{
		DanmakuExchangeHeader: DanmakuExchangeHeader{
			Length:       uint32(length),
			HeaderLength: kHeaderLength,
			ProtocolVer:  ProtoPlainJson,
			Operation:    operation,
			SequenceId:   kSequenceId,
		},
		Body: bodyData,
	}
	return
}
