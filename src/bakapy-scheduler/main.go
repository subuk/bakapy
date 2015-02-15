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

	scriptPool := bakapy.NewDirectoryScriptPool(config)
	metaman := bakapy.NewMetaMan(config)

	var notificators []bakapy.Notificator
	for _, ncConfig := range config.Notificators {
		nc := bakapy.NewScriptedNotificator(scriptPool, ncConfig.Name, ncConfig.Params)
		notificators = append(notificators, nc)
	}

	scheduler := cron.New()
	for jobName, jobConfig := range config.Jobs {
		runSpec := jobConfig.RunAt.SchedulerString()
		logger.Info("adding job %s{%s} to scheduler", jobName, runSpec)

		if jobConfig.Disabled {
			logger.Warning("job %s disabled, skipping", jobName)
			continue
		}
		func(jobName string, jobConfig *bakapy.JobConfig, config *bakapy.Config) {
			scheduler.AddFunc(runSpec, func() {
				logger.Critical("Starting job %s", jobName)
				executor := bakapy.NewBashExecutor(jobConfig.Args, jobConfig.Host, jobConfig.Port, jobConfig.Sudo)
				job := bakapy.NewJob(
					jobName, jobConfig, config.Listen,
					scriptPool, executor,
					metaman,
				)
				err := job.Run()
				if err != nil {
					logger.Warning("job %s failed: %s", jobName, err)
				}
				md, err := metaman.View(job.TaskId)

				if err != nil {
					logger.Critical("cannot get metadata for finished job: %s", err)
					return
				}

				for _, nc := range notificators {
					logger.Debug("executing %s notificator", nc.Name())
					if err := nc.JobFinished(md); err != nil {
						logger.Warning("failed to execute %s notificator: %s", nc.Name(), err)
					} else {
						logger.Debug("notificator %s finished successfully", nc.Name())
					}
				}
			})
		}(jobName, jobConfig, config)
	}

	if *TEST_CONFIG_ONLY {
		return
	}

	scheduler.Start()

}
