package main

import (
	"bakapy"
	"flag"
	"fmt"
	"os"
)

var CONFIG_PATH = flag.String("config", "scheduler.conf", "Path to config file")
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
	metaman := bakapy.NewMetaManClient(config.MetadataAddr, config.Secret)
	spool := bakapy.NewDirectoryScriptPool(config)

	jobName := *JOB_NAME
	jobConfig, jobExist := config.Jobs[jobName]
	if !jobExist {
		fmt.Printf("Job %s not found\n", jobName)
		os.Exit(1)
	}

	notificators := bakapy.NewNotificatorPool()
	for _, ncConfig := range config.Notificators {
		nc := bakapy.NewScriptedNotificator(spool, ncConfig.Name, ncConfig.Params)
		notificators.Add(nc)
	}

	storageAddr, exist := config.Storages[jobConfig.Storage]
	if !exist {
		fmt.Printf("Error: cannot find storage %s definition in config\n", jobConfig.Storage)
		os.Exit(1)
	}
	executor := bakapy.NewBashExecutor(jobConfig.Args, jobConfig.Host, jobConfig.Port, jobConfig.Sudo)
	job := bakapy.NewJob(
		jobName, jobConfig, storageAddr,
		spool, executor, metaman,
		notificators,
	)
	if *FORCE_TASK_ID != "" {
		if len(*FORCE_TASK_ID) != 36 {
			fmt.Println("TaskId length must be 36 bytes")
			os.Exit(1)
		}
		job.TaskId = bakapy.TaskId(*FORCE_TASK_ID)
	}
	job.Run()

	md, err := metaman.View(job.TaskId)
	if !md.Success {
		fmt.Fprintf(os.Stderr, "job failed: %s, %s\n", md.Message, md)
		os.Exit(1)
	} else {
		fmt.Println("Job finished")
	}
}
