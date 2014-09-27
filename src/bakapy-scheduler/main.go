package main

import (
	"bakapy"
	"flag"
	"github.com/op/go-logging"
	"github.com/robfig/cron"
	"os"
)

var logger = logging.MustGetLogger("bakapy.scheduler")
var CONFIG_PATH = flag.String("config", "", "Path to config file")

func setupLogging() {
	format := "%{color}%{time:15:04:05} %{level:.7s} %{module} %{message}%{color:reset}"
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
	logging.SetLevel(logging.DEBUG, "bakapy.scheduler")
}

func makeJobRunner(job *bakapy.Job) func() {
	return func() {
		if job.IsDisabled() {
			logger.Warning("job %s disabled, skipping", job.Name)
			return
		}
		_ = job.Run()
	}
}

func main() {
	flag.Parse()
	setupLogging()

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
		job := bakapy.NewJob(jobName, jobConfig, storage.Tasks)
		scheduler.AddFunc(runSpec, makeJobRunner(job))
	}

	storage.Start()
	scheduler.Start()

	c := make(chan struct{})
	for {
		<-c
	}
}
