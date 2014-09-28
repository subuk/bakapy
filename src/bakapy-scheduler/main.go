package main

import (
	"bakapy"
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"github.com/robfig/cron"
	"os"
)

var logger = logging.MustGetLogger("bakapy.scheduler")
var CONFIG_PATH = flag.String("config", "", "Path to config file")
var LOG_LEVEL = flag.String("loglevel", "info", "Log level")

func main() {
	flag.Parse()
	err := bakapy.SetupLogging(*LOG_LEVEL)
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
		scheduler.AddFunc(runSpec, func() {
			bakapy.RunJob(job, config, logger)
		})
	}

	storage.Start()
	scheduler.Start()

	select {}
}
