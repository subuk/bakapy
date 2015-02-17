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
var CONFIG_PATH = flag.String("config", "bakapy.conf", "Path to config file")
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
	metamanRPC := bakapy.NewMetaRPCServer(metaman)
	bakapy.ServeRPC(config.MetadataListen, config.Secret, metamanRPC)

	notificators := bakapy.NewNotificatorPool()
	for _, ncConfig := range config.Notificators {
		nc := bakapy.NewScriptedNotificator(scriptPool, ncConfig.Name, ncConfig.Params)
		notificators.Add(nc)
	}

	scheduler := cron.New()
	for jobName, jobConfig := range config.Jobs {
		runSpec := jobConfig.RunAt.SchedulerString()

		if jobConfig.Disabled {
			logger.Warning("job %s disabled, skipping", jobName)
			continue
		}
		storageAddr, exist := config.Storages[jobConfig.Storage]
		if !exist {
			logger.Critical("cannot find storage %s definition in config", jobConfig.Storage)
			os.Exit(1)
		}
		executor := bakapy.NewBashExecutor(jobConfig.Args, jobConfig.Host, jobConfig.Port, jobConfig.Sudo)
		job := bakapy.NewJob(
			jobName, jobConfig, storageAddr,
			scriptPool, executor, metaman,
			notificators,
		)
		logger.Info("adding job %s{%s} to scheduler", jobName, runSpec)
		err := scheduler.AddJob(runSpec, job)
		if err != nil {
			logger.Critical("cannot schedule job %s: %s", jobName, err)
			os.Exit(1)
		}
	}

	if *TEST_CONFIG_ONLY {
		return
	}

	scheduler.Start()
	<-(make(chan int))
}
