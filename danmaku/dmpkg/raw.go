package dmpkg

import (
	"encoding/json"
	"fmt"
	"math"
)

const MaxBodyLength = math.MaxUint32 - uint64(HeaderLength)

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

	length := uint64(HeaderLength + len(bodyData))
	if length > MaxBodyLength {
		err = fmt.Errorf("body is too large (> %d)", MaxBodyLength)
		return
	}
	exc = DanmakuExchange{
		DanmakuExchangeHeader: DanmakuExchangeHeader{
			Length:       uint32(length),
			HeaderLength: HeaderLength,
			ProtocolVer:  ProtoMinimal,
			Operation:    operation,
			SequenceId:   SequenceId,
		},
		Body: bodyData,
	}
	return
}
