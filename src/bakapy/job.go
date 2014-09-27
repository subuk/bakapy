package bakapy

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/op/go-logging"
	"net"
	"os"
	"time"
)

type Job struct {
	Name     string
	storChan StorageTaskChan
	cfg      JobConfig
	logger   *logging.Logger
}

type TaskId string

type JobMetadata struct {
	Pid        int
	Addr       net.Addr
	Config     JobConfig
	StartTime  time.Time
	EndTime    time.Time
	ExpireTime time.Time
	TaskId     TaskId
	Files      []string
	Success    bool
	Message    string
	TotalSize  uint
}

func NewJob(name string, cfg JobConfig, storChan StorageTaskChan) *Job {
	loggerName := fmt.Sprintf("bakapy.job[%s]", name)
	return &Job{
		Name:     name,
		cfg:      cfg,
		logger:   logging.MustGetLogger(loggerName),
		storChan: storChan,
	}
}

func (job *Job) Run() *JobMetadata {
	job.logger.Info("starting up")
	metadata := &JobMetadata{
		Pid:       os.Getpid(),
		Config:    job.cfg,
		StartTime: time.Now(),
	}
	metadata.TaskId = TaskId(uuid.NewUUID())
	// TODO:
	// * read shell script
	// * prepend shell script header
	// * send job metadata to storage
	// * start task over ssh
	return metadata
}

func (job *Job) IsDisabled() bool {
	return job.cfg.Disabled
}
