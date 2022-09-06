package dmpkg

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/lunixbochs/struc"
	"io"
)

type DanmakuExchangeHeader struct {
	// Length total remaining bytes of this exchange, excluding `Length` itself
	Length uint32
	// HeaderLength = Length - len(Body) + 4, always equals to 16
	HeaderLength uint16
	ProtocolVer  ProtocolVer
	Operation    Operation
	// SequenceId is always 1
	SequenceId uint32
}

// DanmakuExchange represents an actual message sent from client or server. This is an atomic unit.
type DanmakuExchange struct {
	DanmakuExchangeHeader
	Body []byte
}

func (e *DanmakuExchange) String() string {
	return fmt.Sprintf("DanmakuExchange{length=%v, protocol=%v, operation=%v, body=%v}",
		e.Length, e.ProtocolVer, e.Operation, e.Body)
}

const kHeaderLength = 16
const kSequenceId = 1

type ProtocolVer uint16

const (
	// ProtoPlainJson the body is plain JSON text
	ProtoPlainJson ProtocolVer = 0
	// ProtoMinimal the body is uint32 watcher count (big endian)
	ProtoMinimal ProtocolVer = 1
	// ProtoZlib the body is a zlib compressed package
	ProtoZlib ProtocolVer = 2
	// ProtoBrotli the body is a brotli compressed package
	ProtoBrotli ProtocolVer = 3
)

const kUidGuest = 0
const kPlatformWeb = "web"
const kAuthTypeDefault = 2 // magic number, not sure what does it mean

func (e *DanmakuExchange) Marshal() (data []byte, err error) {
	var buffer bytes.Buffer
	// only unpack header with struc, since it does not support indirect variable field length calculation
	err = struc.Pack(&buffer, &e.DanmakuExchangeHeader)
	if err != nil {
		err = fmt.Errorf("cannot pack an exchange into binary form: %w", err)
		return
	}
	data = buffer.Bytes()
	data = append(data, e.Body...)
	return
}

// Inflate decompresses the body if it is compressed
func (e *DanmakuExchange) Inflate() (ret DanmakuExchange, err error) {
	switch e.ProtocolVer {
	case ProtoMinimal:
		fallthrough
	case ProtoPlainJson:
		ret = *e
	case ProtoBrotli:
		var data []byte
		rd := brotli.NewReader(bytes.NewReader(e.Body))
		data, err = io.ReadAll(rd)
		if err != nil {
			err = fmt.Errorf("cannot decompress exchange body: %w", err)
			return
		}
		var nestedExchange DanmakuExchange
		nestedExchange, err = DecodeExchange(data)
		if err != nil {
			err = fmt.Errorf("cannot decode nested exchange: %w", err)
			return
		}
		return nestedExchange.Inflate()
	case ProtoZlib:
		var data []byte
		var rd io.ReadCloser
		rd, err = zlib.NewReader(bytes.NewReader(e.Body))
		if err != nil {
			err = fmt.Errorf("cannot create zlib reader: %w", err)
			return
		}
		data, err = io.ReadAll(rd)
		if err != nil {
			err = fmt.Errorf("cannot decompress exchange body: %w", err)
			return
		}
		var nestedExchange DanmakuExchange
		nestedExchange, err = DecodeExchange(data)
		if err != nil {
			err = fmt.Errorf("cannot decode nested exchange: %w", err)
			return
		}
		return nestedExchange.Inflate()
	}
	return
}
