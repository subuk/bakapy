package main

import (
	"bakapy"
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"os"
)

var logger = logging.MustGetLogger("bakapy.scheduler")
var CONFIG_PATH = flag.String("config", "/etc/bakapy/bakapy.conf", "Path to config file")
var LOG_LEVEL = flag.String("loglevel", "debug", "Log level")
var JOB_NAME = flag.String("job", "REQUIRED", "Job name")

func main() {
	flag.Parse()
	err := bakapy.SetupLogging(*LOG_LEVEL)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	config, err := bakapy.ParseConfig(*CONFIG_PATH)
	if err != nil {
		logger.Fatal(err)
	}

	storage := bakapy.NewStorage(config)

	jobName := *JOB_NAME
	jobConfig, jobExist := config.Jobs[jobName]
	if !jobExist {
		logger.Fatalf("Job %s not found", jobName)
		return
	}

	job := bakapy.NewJob(
		jobName, jobConfig,
		storage.JobsChan, config,
	)
	storage.Start()
	bakapy.RunJob(job, config, logger)
	storage.Wait()
}
