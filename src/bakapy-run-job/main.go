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
var FORCE_TASK_ID = flag.String("taskid", "", "Use this task id for job")

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
	storage := bakapy.NewStorage(config, metaman)

	jobName := *JOB_NAME
	jobConfig, jobExist := config.Jobs[jobName]
	if !jobExist {
		fmt.Printf("Job %s not found\n", jobName)
		os.Exit(1)
	}

	storage.Start()

	executor := bakapy.NewBashExecutor(jobConfig.Args, jobConfig.Host, jobConfig.Port, jobConfig.Sudo)
	job := bakapy.NewJob(jobName, jobConfig, config.Listen, config.CommandDir, executor, metaman)
	if *FORCE_TASK_ID != "" {
		if len(*FORCE_TASK_ID) != 36 {
			fmt.Println("TaskId length must be 36 bytes")
			os.Exit(1)
		}
		job.TaskId = bakapy.TaskId(*FORCE_TASK_ID)
	}
	if err := job.Run(); err != nil {
		fmt.Printf("Job failed: %s", err)
		os.Exit(1)
	}
	fmt.Println("Job finished")
}
