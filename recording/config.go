package recording

import (
	"bilibili-livestream-archiver/common"
	"fmt"
)

type TaskConfig struct {
	RoomId    common.RoomId   `mapstructure:"room_id"`
	Transport TransportConfig `mapstructure:"transport"`
	Download  DownloadConfig  `mapstructure:"download"`
}

type TransportConfig struct {
	SocketTimeoutSeconds int `mapstructure:"socket_timeout_seconds"`
	RetryIntervalSeconds int `mapstructure:"retry_interval_seconds"`
	MaxRetryTimes        int `mapstructure:"max_retry_times"`
}

type DownloadConfig struct {
	SaveDirectory string `mapstructure:"save_directory"`
}

func DefaultTransportConfig() TransportConfig {
	return TransportConfig{
		SocketTimeoutSeconds: 10,
		RetryIntervalSeconds: 2,
		MaxRetryTimes:        5,
	}
}

func (t TaskConfig) String() string {
	return fmt.Sprintf("room: %v, %v, %v", t.RoomId, t.Transport.String(), t.Download.String())
}

func (t TransportConfig) String() string {
	return fmt.Sprintf("socket timeout: %vs, retry interval: %vs, max retry times: %v",
		t.SocketTimeoutSeconds, t.RetryIntervalSeconds, t.MaxRetryTimes)
}

func (d DownloadConfig) String() string {
	return fmt.Sprintf("save directory: %v", d.SaveDirectory)
}
