package main

import (
	"github.com/keuin/slbr/api/agent"
	"github.com/keuin/slbr/recording"
	"github.com/samber/lo"
)

type agentImpl struct {
	tasks *[]*recording.RunningTask
}

func (a *agentImpl) GetTasks() []agent.TaskInfo {
	return lo.Map(*a.tasks, func(t *recording.RunningTask, _ int) agent.TaskInfo {
		return agent.TaskInfo{
			LiveRoom: agent.LiveRoomInfo{
				ID:    t.RoomId,
				Title: t.GetRoomTitle(),
			},
			Status: t.GetStatus().String(),
		}
	})
}
