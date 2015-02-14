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

	metaman := bakapy.NewMetaMan(config)
	storage := bakapy.NewStorage(config, metaman)
	sender := bakapy.NewMailSender(config.SMTP)

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
				executor := bakapy.NewBashExecutor(jobConfig.Args, jobConfig.Host, jobConfig.Port, jobConfig.Sudo)
				job := bakapy.NewJob(
					jobName, jobConfig, config.Listen,
					config.CommandDir, executor,
					metaman,
				)
				err := job.Run()
				if err != nil {
					logger.Critical("Job %s failed: %s", job.TaskId, err)
					md, err := metaman.View(job.TaskId)
					if err != nil {
						logger.Critical("Cannot get metadata for finished job: %s", err)
						return
					}
					if err := sender.SendFailedJobNotification(&md); err != nil {
						logger.Critical("cannot send failed job notification: %s", err)
					}

				}
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
