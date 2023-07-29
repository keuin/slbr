package main

import "github.com/keuin/slbr/recording"

type GlobalConfig struct {
	Tasks     []recording.TaskConfig `mapstructure:"tasks"`
	ApiServer string                 `mapstructure:"api_server"`
}
