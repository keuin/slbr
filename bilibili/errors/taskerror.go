package errors

import (
	"fmt"
	"github.com/samber/lo"
	"strings"
)

type Type int

const (
	// GetRoomInfo means failed to read live room information
	GetRoomInfo Type = iota
	// GetLiveInfo means failed to get live information
	GetLiveInfo
	// StreamCopy means an error occurred while reading stream video data
	StreamCopy
	// LiveEnded means the live is ended
	LiveEnded
	// DanmakuServerConnection means failed to connect to danmaku server, which provides danmaku and other control data
	DanmakuServerConnection
	// Heartbeat means an error occurred while sending heartbeat message, which may indicate a broken network connection
	Heartbeat
	// InitialLiveStatus means failed to get the live status for the first time, which happens in the early stage
	InitialLiveStatus
	// DanmakuExchangeRead means an error occurred while reading danmaku datagram from the server,
	// which may indicate a broken network connection
	DanmakuExchangeRead
	// GetDanmakuServerInfo means failed to get danmaku server info
	GetDanmakuServerInfo
	// RecoverLiveStatusChecker means failed to restart live status checker
	RecoverLiveStatusChecker

	// FileCreation means failed to create a file
	FileCreation
	// InvalidLiveInfo means the live info is insufficient for recording
	InvalidLiveInfo
	// LiveStatusWatch means the live status watcher encountered an unrecoverable error
	LiveStatusWatch
	// Unknown means the error type is unexpected and is not suitable to retry for safety
	Unknown
	// InvalidAuthProtocol means authentication failed because the protocol is invalid,
	// which may indicate the protocol implementation is outdated
	InvalidAuthProtocol
	// MessageDecompression means the message cannot be decompressed, and we cannot understand its content
	MessageDecompression
	// JsonDecode means we cannot decode a datum which is expected to be a JSON object string
	JsonDecode
)

var recoverableErrors = []Type{
	GetRoomInfo,
	GetLiveInfo,
	StreamCopy,
	LiveEnded,
	DanmakuServerConnection,
	Heartbeat,
	InitialLiveStatus,
	DanmakuExchangeRead,
	GetDanmakuServerInfo,
	RecoverLiveStatusChecker,
}

var errorStrings = map[Type]string{
	GetRoomInfo:              "failed to get living room information",
	GetLiveInfo:              "failed to get live info",
	StreamCopy:               "stream copy was unexpectedly interrupted",
	LiveEnded:                "live is ended",
	DanmakuServerConnection:  "failed to connect to danmaku server",
	Heartbeat:                "heartbeat failed",
	InitialLiveStatus:        "check initial live status failed",
	DanmakuExchangeRead:      "failed to read exchange from server",
	GetDanmakuServerInfo:     "cannot get notification server info",
	RecoverLiveStatusChecker: "when recovering from a previous error, another error occurred",
	FileCreation:             "failed to create file",
	InvalidLiveInfo:          "invalid live info",
	LiveStatusWatch:          "failed to watch live status",
	Unknown:                  "unexpected error type",
	InvalidAuthProtocol:      "authentication failed, invalid protocol",
	MessageDecompression:     "failed to decompress server message",
	JsonDecode:               "invalid JSON response from server",
}

func (t Type) String() string {
	if s, ok := errorStrings[t]; ok {
		return s
	}
	return fmt.Sprintf("<Type %v>", int(t))
}

type taskError struct {
	typ Type
	err []error
}

func (e *taskError) Message() string {
	//TODO implement me
	panic("implement me")
}

func (e *taskError) IsRecoverable() bool {
	return lo.Contains(recoverableErrors, e.typ)
}

func (e *taskError) Unwrap() []error {
	return e.err
}

func (e taskError) Error() string {
	sb := strings.Builder{}
	if e.IsRecoverable() {
		sb.WriteString("recoverable task error")
	} else {
		sb.WriteString("unrecoverable task error")
	}
	sb.WriteString(": ")
	sb.WriteString(fmt.Sprintf("%v", e.typ))
	if len(e.err) > 0 {
		sb.WriteString(", ")
		for i := range e.err {
			sb.WriteString(fmt.Sprintf("%v", e.err[i]))
			if i != len(e.err)-1 {
				sb.WriteString(", ")
			}
		}
	}
	return sb.String()
}

type TaskError interface {
	// IsRecoverable reports if this task error is safe to retry.
	IsRecoverable() bool
	// Unwrap returns the underneath errors which cause the task error.
	Unwrap() []error
	// Message returns the detailed task-specific error message.
	Message() string
	// Error returns this error as string
	Error() string
}

func NewError(typ Type, err ...error) TaskError {
	return &taskError{
		typ: typ,
		err: err,
	}
}
