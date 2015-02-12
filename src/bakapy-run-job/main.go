package main

import (
	"bakapy"
	"flag"
	"fmt"
	"os"
)

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
		fmt.Println(err.Error())
		os.Exit(1)
	}

	metaman := bakapy.NewMetaMan(config)
	storage := bakapy.NewStorage(config)

	jobName := *JOB_NAME
	jobConfig, jobExist := config.Jobs[jobName]
	if !jobExist {
		fmt.Printf("Job %s not found\n", jobName)
		os.Exit(1)
	}

	storage.Start()
	bakapy.RunJob(jobName, jobConfig, config, storage, metaman)
}
