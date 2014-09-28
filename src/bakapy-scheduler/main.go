package main

import (
	"bakapy"
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"github.com/robfig/cron"
	"os"
	"path"
	"strings"
)

var logger = logging.MustGetLogger("bakapy.scheduler")
var CONFIG_PATH = flag.String("config", "", "Path to config file")
var LOG_LEVEL = flag.String("loglevel", "info", "Log level")

func setupLogging() error {
	format := "%{color}%{time:15:04:05} %{level:.8s} %{module} %{message}%{color:reset}"
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
	level, err := logging.LogLevel(strings.ToUpper(*LOG_LEVEL))
	if err != nil {
		return err
	}
	logging.SetLevel(level, "")
	return nil
}

func makeJobRunner(job *bakapy.Job, config *bakapy.Config) func() {
	return func() {
		if job.IsDisabled() {
			logger.Warning("job %s disabled, skipping", job.Name)
			return
		}
		metadata := job.Run()
		saveTo := path.Join(config.MetadataDir, string(metadata.TaskId))
		err := metadata.Save(saveTo)
		if err != nil {
			logger.Critical("cannot save metadata: %s", err)
		}
		logger.Info("metadata for job %s successfully saved to %s", metadata.TaskId, saveTo)
		if !metadata.Success {
			logger.Critical("job '%s' failed", job.Name)
		} else {
			logger.Info("job '%s' finished", job.Name)
		}
	}
}

func main() {
	flag.Parse()
	err := setupLogging()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	config, err := bakapy.ParseConfig(*CONFIG_PATH)
	if err != nil {
		logger.Fatal("Error: ", err)
	}

	logger.Debug(string(config.PrettyFmt()))

	storage := bakapy.NewStorage(config)

	scheduler := cron.New()
	for jobName, jobConfig := range config.Jobs {
		runSpec := jobConfig.RunAt.SchedulerString()
		logger.Info("adding job %s{%s} to scheduler", jobName, runSpec)
		job := bakapy.NewJob(
			jobName, jobConfig,
			storage.JobsChan, config,
		)
		scheduler.AddFunc(runSpec, makeJobRunner(job, config))
	}

	storage.Start()
	scheduler.Start()

	c := make(chan struct{})
	for {
		<-c
	}
}
