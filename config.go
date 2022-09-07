package main

import "bilibili-livestream-archiver/recording"

type GlobalConfig struct {
	Tasks []recording.TaskConfig `mapstructure:"tasks"`
}
