package main

import (
	"bakapy"
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"github.com/robfig/cron"
	"os"
	"time"
)

var logger = logging.MustGetLogger("bakapy.scheduler")
var CONFIG_PATH = flag.String("config", "/etc/bakapy/bakapy.conf", "Path to config file")
var LOG_LEVEL = flag.String("loglevel", "debug", "Log level")
var TEST_CONFIG_ONLY = flag.Bool("test", false, "Check config and exit")

func main() {
	flag.Parse()
	err := bakapy.SetupLogging(*LOG_LEVEL)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	config, err := bakapy.ParseConfig(*CONFIG_PATH)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %s\n", err)
		os.Exit(1)
	}

	logger.Debug(string(config.PrettyFmt()))

	storage := bakapy.NewStorage(config)

	scheduler := cron.New()
	for jobName, jobConfig := range config.Jobs {
		runSpec := jobConfig.RunAt.SchedulerString()
		logger.Info("adding job %s{%s} to scheduler", jobName, runSpec)

		if jobConfig.Disabled {
			logger.Warning("job %s disabled, skipping", jobName)
			continue
		}
		func(jobName string, jobConfig *bakapy.JobConfig, config *bakapy.Config, storage *bakapy.Storage) {
			scheduler.AddFunc(runSpec, func() {
				logger.Critical("Starting job %s", jobName)
				bakapy.RunJob(jobName, jobConfig, config, storage)
			})
		}(jobName, jobConfig, config, storage)
	}

	if *TEST_CONFIG_ONLY {
		return
	}

	storage.Start()
	scheduler.Start()

	for {
		err := storage.CleanupExpired()
		if err != nil {
			logger.Warning("cleanup failed: %s", err.Error())
		}
		time.Sleep(time.Minute)
	}

}
