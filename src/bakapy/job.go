package bakapy

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/op/go-logging"
	"os"
	"path"
	"strings"
	"time"
)

type TaskId string

func (t *TaskId) String() string {
	return string(*t)
}

type JobTemplateContext struct {
	Job              *Job
	FILENAME_LEN_LEN uint
}

func (jctx *JobTemplateContext) ToHost() string {
	return strings.Split(jctx.Job.StorageAddr, ":")[0]
}

func (jctx *JobTemplateContext) ToPort() string {
	return strings.Split(jctx.Job.StorageAddr, ":")[1]
}

type Job struct {
	Name        string
	TaskId      TaskId
	StorageAddr string
	CommandDir  string
	executor    Executer
	cfg         *JobConfig
	logger      *logging.Logger
	metaman     MetaManager
}

func NewJob(name string, cfg *JobConfig, StorageAddr string, commandDir string, executor Executer, metaman MetaManager) *Job {
	taskId := TaskId(uuid.NewUUID().String())
	loggerName := fmt.Sprintf("bakapy.job[%s][%s]", name, taskId)
	return &Job{
		Name:        name,
		TaskId:      taskId,
		StorageAddr: StorageAddr,
		CommandDir:  commandDir,
		cfg:         cfg,
		executor:    executor,
		logger:      logging.MustGetLogger(loggerName),
		metaman:     metaman,
	}
}

func (job *Job) getScript() ([]byte, error) {
	script := new(bytes.Buffer)
	err := JOB_TEMPLATE.Execute(script, &JobTemplateContext{
		Job:              job,
		FILENAME_LEN_LEN: STORAGE_FILENAME_LEN_LEN,
	})
	if err != nil {
		return nil, err
	}

	scriptPath := path.Join(job.CommandDir, job.cfg.Command)
	job.logger.Debug("reading command file %s", scriptPath)
	fd, err := os.Open(scriptPath)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	_, err = script.ReadFrom(fd)
	if err != nil {
		return nil, err
	}
	return script.Bytes(), nil
}

func (job *Job) Run() error {
	job.logger.Info("starting up")
	err := job.metaman.Add(
		job.Name,
		job.cfg.Namespace,
		job.cfg.Command,
		job.TaskId,
		job.cfg.Gzip,
		job.cfg.MaxAge,
	)
	if err != nil {
		return fmt.Errorf("cannot add metadata: %s", err)
	}
	script, err := job.getScript()
	if err != nil {
		job.logger.Warning("cannot get job script: %s", err.Error())
		job.metaman.Update(job.TaskId, func(md *Metadata) {
			md.Message = err.Error()
		})
		return err
	}

	output := new(bytes.Buffer)
	errput := new(bytes.Buffer)
	err = job.executor.Execute(script, output, errput)

	job.logger.Debug("Command output: %s", output.String())
	job.logger.Debug("Command errput: %s", errput.String())

	job.metaman.Update(job.TaskId, func(md *Metadata) {
		md.Output = output.Bytes()
		md.Errput = errput.Bytes()
	})

	if err != nil {
		job.logger.Warning("command failed: %s", err)
		job.metaman.Update(job.TaskId, func(md *Metadata) {
			md.Success = false
			md.Message = err.Error()
			md.EndTime = time.Now().UTC()
		})
		return err
	}

	job.metaman.Update(job.TaskId, func(md *Metadata) {
		md.Success = true
		md.Message = "OK"
		md.EndTime = time.Now().UTC()
	})
	return nil
}
