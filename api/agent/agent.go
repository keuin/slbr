package agent

import "github.com/keuin/slbr/types"

type Agent interface {
	GetTasks() []TaskInfo
}

type TaskStatus string

type LiveRoomInfo struct {
	ID    types.RoomId `json:"id"`
	Title *string      `json:"title"`
}

type TaskInfo struct {
	LiveRoom LiveRoomInfo `json:"live_room"`
	Status   string       `json:"status"`
}
