package dmpkg

import (
	"bytes"
	"fmt"
	"github.com/lunixbochs/struc"
)

func DecodeExchange(data []byte) (exc DanmakuExchange, err error) {
	if ln := len(data); ln < kHeaderLength {
		err = fmt.Errorf("incomplete datagram: length = %v < %v", ln, kHeaderLength)
		return
	}

	// unpack header
	var exchangeHeader DanmakuExchangeHeader
	err = struc.Unpack(bytes.NewReader(data[:kHeaderLength]), &exchangeHeader)
	if err != nil {
		err = fmt.Errorf("cannot unpack exchange header: %w", err)
		return
	}
	headerLength := exchangeHeader.HeaderLength

	// validate header length, fail fast if not match
	if headerLength != kHeaderLength {
		err = fmt.Errorf("invalid header length, "+
			"the protocol implementation might be obsolete: %v != %v", headerLength, kHeaderLength)
		return
	}

	// special process
	// TODO decouple this
	// The server OpHeartbeatAck contains an extra 4-bytes header entry in the body, maybe a heat value
	var body []byte
	// copy body
	body = make([]byte, exchangeHeader.Length-uint32(headerLength))
	copy(body, data[headerLength:])

	exc = DanmakuExchange{
		DanmakuExchangeHeader: exchangeHeader,
		Body:                  body,
	}
	return
}
