package recording

import (
	"bilibili-livestream-archiver/common"
	"fmt"
)

type TaskConfig struct {
	RoomId    common.RoomId   `mapstructure:"room_id"`
	Transport TransportConfig `mapstructure:"transport"`
	Download  DownloadConfig  `mapstructure:"download"`
	Watch     WatchConfig     `mapstructure:"watch"`
}

type TransportConfig struct {
	SocketTimeoutSeconds int `mapstructure:"socket_timeout_seconds"`
	RetryIntervalSeconds int `mapstructure:"retry_interval_seconds"`
	MaxRetryTimes        int `mapstructure:"max_retry_times"`
}

type DownloadConfig struct {
	SaveDirectory        string `mapstructure:"save_directory"`
	DiskWriteBufferBytes int    `mapstructure:"disk_write_buffer_bytes"`
}

type WatchConfig struct {
	LiveInterruptedRestartSleepSeconds int `mapstructure:"live_interrupted_restart_sleep_seconds"`
}

func DefaultTransportConfig() TransportConfig {
	return TransportConfig{
		SocketTimeoutSeconds: 10,
		RetryIntervalSeconds: 2,
		MaxRetryTimes:        5,
	}
}

func (t TaskConfig) String() string {
	return fmt.Sprintf("Room ID: %v, %v, %v", t.RoomId, t.Transport.String(), t.Download.String())
}

func (t TransportConfig) String() string {
	return fmt.Sprintf("Socket timeout: %vs, Retry interval: %vs, Max retry times: %v",
		t.SocketTimeoutSeconds, t.RetryIntervalSeconds, t.MaxRetryTimes)
}

func (d DownloadConfig) String() string {
	return fmt.Sprintf("Save directory: \"%v\"", d.SaveDirectory)
}
