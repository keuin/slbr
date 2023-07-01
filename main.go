package main

/*
In this file we implement config file and command line arguments parsing.
Task lifecycle management are implemented in recording package.
*/

import (
	"context"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/keuin/slbr/bilibili"
	"github.com/keuin/slbr/common"
	"github.com/keuin/slbr/logging"
	"github.com/keuin/slbr/recording"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/mo"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
)

const defaultDiskBufSize = uint64(1024 * 1024) // 1MiB

var globalConfig *GlobalConfig

func getTasks() (tasks []recording.TaskConfig) {
	var err error
	parser := argparse.NewParser(
		"slbr",
		"Record bilibili live streams",
	)
	defer func() {
		if err != nil {
			fmt.Printf("ERROR: %v.\n", err)
			fmt.Print(parser.Usage(""))
			os.Exit(0)
		}
	}()
	configFilePtr := parser.String(
		"c", "config",
		&argparse.Options{
			Required: false,
			Help:     "Specify which configuration file to use. JSON, TOML and YAML are all supported",
		},
	)
	rooms := parser.IntList(
		"s", "room",
		&argparse.Options{
			Required: false,
			Help: "Specify which room to record. " +
				"The ID is included in bilibili live webpage url. " +
				"Set this to run without config file",
		},
	)
	saveToPtr := parser.String(
		"o", "save-to",
		&argparse.Options{
			Required: false,
			Help: "Specify the directory where to save records. " +
				"If not set, process working directory is used",
		},
	)
	diskBufSizePtr := parser.Int(
		"b", "disk-write-buffer",
		&argparse.Options{
			Required: false,
			Help: "Specify disk write buffer size (bytes). " +
				"The real minimum buffer size is determined by OS. " +
				"Setting this to a large value may make stopping take a long time",
			Default: 4194304,
		},
	)

	err = parser.Parse(os.Args)
	if err != nil {
		return
	}

	fromCli := len(*rooms) > 0
	fromFile := *configFilePtr != ""

	if fromCli && fromFile {
		err = fmt.Errorf("cannot specify room id argument and config file at the same time")
		return
	}

	if !fromCli && !fromFile {
		err = fmt.Errorf("no task specified")
		return
	}

	if fromFile {
		configFile := *configFilePtr
		fmt.Printf("Config file: %v\n", configFile)

		viper.SetConfigFile(configFile)

		err = viper.ReadInConfig()
		if err != nil {
			err = fmt.Errorf("cannot open config file \"%v\": %w", configFile, err)
			return
		}

		if err != nil {
			err = fmt.Errorf("cannot read config file \"%v\": %w", configFile, err)
			return
		}
		var gc GlobalConfig
		netType := reflect.TypeOf(bilibili.IP64)
		err = viper.Unmarshal(&gc, func(conf *mapstructure.DecoderConfig) {
			conf.DecodeHook = func(from reflect.Value, to reflect.Value) (interface{}, error) {
				if to.Type() == netType &&
					bilibili.IpNetType(from.String()).GetDialNetString() == "" {
					return nil, fmt.Errorf("invalid IpNetType: %v", from.String())
				}
				return from.Interface(), nil
			}
		})
		if err != nil {
			err = fmt.Errorf("cannot parse config file \"%v\": %w", configFile, err)
			return
		}
		globalConfig = &gc
		return globalConfig.Tasks
	}

	// generate task list from cli
	taskCount := len(*rooms)
	tasks = make([]recording.TaskConfig, taskCount)
	saveTo := mo.EmptyableToOption(*saveToPtr).OrElse(".")
	diskBufSize := uint64(*diskBufSizePtr)
	if *diskBufSizePtr <= 0 {
		diskBufSize = defaultDiskBufSize
	}
	for i := 0; i < taskCount; i++ {
		tasks[i] = recording.TaskConfig{
			RoomId:    common.RoomId((*rooms)[i]),
			Transport: recording.DefaultTransportConfig(),
			Download: recording.DownloadConfig{
				DiskWriteBufferBytes: int64(diskBufSize),
				SaveDirectory:        saveTo,
			},
		}
	}

	return
}

func main() {
	logger := log.Default()
	taskConfigs := getTasks()
	tasks := make([]recording.RunningTask, len(taskConfigs))

	wg := sync.WaitGroup{}
	ctxTasks, cancelTasks := context.WithCancel(context.Background())
	fmt.Println("Record tasks:")
	for i, task := range taskConfigs {
		tasks[i] = recording.NewRunningTask(
			taskConfigs[i],
			ctxTasks,
			func() {},
			func() { wg.Done() },
			logging.NewWrappedLogger(logger, fmt.Sprintf("room %v", task.RoomId)),
		)
		fmt.Printf("[%2d] %s\n", i+1, task)
	}
	fmt.Println("")

	logger.Printf("Starting tasks...")

	for i := range tasks {
		wg.Add(1)
		err := tasks[i].StartTask()
		if err != nil {
			logger.Printf("Cannot start task %v (room %v): %v. Skip.", i, tasks[i].RoomId, err)
			wg.Done()
		}
	}

	// listen on stop signals
	chSigStop := make(chan os.Signal)
	signal.Notify(chSigStop,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM)

	chSigQuit := make(chan os.Signal)
	signal.Notify(chSigQuit, syscall.SIGQUIT)
	go func() {
		select {
		case <-chSigStop:
			logger.Println("Stopping all tasks...")
			cancelTasks()
		case <-chSigQuit:
			logger.Println("Aborted.")
			os.Exit(0)
		}
	}()

	// block main goroutine on task goroutines
	defer func() {
		wg.Wait()
		logger.Println("YABR is stopped.")
	}()
}
