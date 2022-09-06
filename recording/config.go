package recording

import "bilibili-livestream-archiver/common"

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
	SaveDirectory    string `mapstructure:"save_directory"`
	FileNameTemplate string `mapstructure:"file_name_template"`
}
