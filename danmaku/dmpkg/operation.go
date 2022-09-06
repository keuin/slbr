package dmpkg

type Operation uint32

// Operations
const (
	OpHeartbeat    Operation = 2
	OpHeartbeatAck Operation = 3
	OpLayer7Data   Operation = 5
	OpConnect      Operation = 7
	OpConnectOk    Operation = 8
)

var opStringMap = map[Operation]string{
	OpHeartbeat:    "HeartBeat",
	OpHeartbeatAck: "HeartBeatAck",
	OpLayer7Data:   "AppData",
	OpConnect:      "Connect",
	OpConnectOk:    "ConnectOk",
}

func (o Operation) String() string {
	s, exists := opStringMap[o]
	if !exists {
		return "<Unknown>"
	}
	return s
}
