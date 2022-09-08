package main

import (
	"bilibili-livestream-archiver/common"
	"bilibili-livestream-archiver/recording"
	"context"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

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
			Help:     "Specify which configuration file to use. JSON, TOML and YAML are all supported.",
		},
	)
	rooms := parser.IntList(
		"s", "room",
		&argparse.Options{
			Required: false,
			Help:     "The room id to record. Set this to run without config file",
		},
	)
	saveToPtr := parser.String(
		"o", "save-to",
		&argparse.Options{
			Required: false,
			Help:     "Specify which configuration file to use",
			Default:  ".",
		},
	)
	diskBufSizePtr := parser.Int(
		"b", "disk-write-buffer",
		&argparse.Options{
			Required: false,
			Help:     "Specify disk write buffer size (bytes). The real minimum buffer size is determined by OS",
			Default:  -1,
		},
	)

	err = parser.Parse(os.Args)
	if err != nil {
		return
	}

	fromCli := len(*rooms) > 0
	fromFile := *configFilePtr != ""

	if fromCli == fromFile {
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
		err = viper.Unmarshal(&gc)
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
	saveTo := *saveToPtr
	diskBufSize := *diskBufSizePtr
	for i := 0; i < taskCount; i++ {
		tasks[i] = recording.TaskConfig{
			RoomId:    common.RoomId((*rooms)[i]),
			Transport: recording.DefaultTransportConfig(),
			Download: recording.DownloadConfig{
				DiskWriteBufferBytes: diskBufSize,
				SaveDirectory:        saveTo,
			},
		}
	}

	return
}

func main() {
	tasks := getTasks()

	fmt.Println("Record tasks:")
	for i, task := range tasks {
		fmt.Printf("[%2d] %s\n", i+1, task)
	}
	fmt.Println("")

	logger := log.Default()

	logger.Printf("Starting tasks...")
	wg := sync.WaitGroup{}
	defer func() {
		wg.Wait()
		logger.Println("Stopping YABR...")
	}()
	ctx, cancelTasks := context.WithCancel(context.Background())
	for _, task := range tasks {
		wg.Add(1)
		go recording.RunTask(ctx, &wg, &task)
	}

	// listen on stop signals
	chSigStop := make(chan os.Signal)
	signal.Notify(chSigStop,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM)
	go func() {
		<-chSigStop
		cancelTasks()
	}()

	chSigQuit := make(chan os.Signal)
	signal.Notify(chSigQuit, syscall.SIGQUIT)
	go func() {
		<-chSigQuit
		os.Exit(0)
	}()

	// block main goroutine on task goroutines
	wg.Wait()
}
