package main

import (
	"bilibili-livestream-archiver/recording"
	"context"
	"fmt"
)

func main() {
	task := recording.TaskConfig{
		RoomId: 7777,
		Transport: recording.TransportConfig{
			SocketTimeoutSeconds: 10,
			RetryIntervalSeconds: 5,
			MaxRetryTimes:        5,
		},
		Download: recording.DownloadConfig{
			SaveDirectory:    ".",
			FileNameTemplate: "",
		},
	}
	chResult := make(chan recording.TaskResult)
	go recording.RunTask(
		context.Background(),
		&task,
		chResult,
	)
	result := <-chResult
	fmt.Println(result.Error)
}
