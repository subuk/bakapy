package bakapy

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/op/go-logging"
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
	executor    Executer
	scripts     BackupScriptPool
	cfg         *JobConfig
	logger      *logging.Logger
	metaman     MetaManager
}

func NewJob(name string, cfg *JobConfig, StorageAddr string, scripts BackupScriptPool, executor Executer, metaman MetaManager) *Job {
	taskId := TaskId(uuid.NewUUID().String())
	loggerName := fmt.Sprintf("bakapy.job[%s][%s]", name, taskId)
	return &Job{
		Name:        name,
		TaskId:      taskId,
		StorageAddr: StorageAddr,
		scripts:     scripts,
		cfg:         cfg,
		executor:    executor,
		logger:      logging.MustGetLogger(loggerName),
		metaman:     metaman,
	}
}

func (job *Job) getScript() ([]byte, error) {
	fullScript := new(bytes.Buffer)
	err := JOB_TEMPLATE.Execute(fullScript, &JobTemplateContext{
		Job:              job,
		FILENAME_LEN_LEN: STORAGE_FILENAME_LEN_LEN,
	})
	if err != nil {
		return nil, err
	}

	script, err := job.scripts.BackupScript(job.cfg.Command)
	if err != nil {
		return nil, fmt.Errorf("cannot find backup script %s: %s", job.cfg.Command, err)
	}

	_, err = fullScript.ReadFrom(bytes.NewReader(script))
	if err != nil {
		return nil, err
	}
	return fullScript.Bytes(), nil
}

func (job *Job) Run() error {
	job.logger.Info("starting up")
	now := time.Now().UTC()
	err := job.metaman.Add(job.TaskId, Metadata{
		JobName:    job.Name,
		Gzip:       job.cfg.Gzip,
		Namespace:  job.cfg.Namespace,
		Command:    job.cfg.Command,
		StartTime:  now,
		ExpireTime: now.Add(job.cfg.MaxAge),
	})
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
	jobErr := job.executor.Execute(script, output, errput)

	job.logger.Debug("Command output: %s", output.String())
	job.logger.Debug("Command errput: %s", errput.String())

	if jobErr != nil {
		mdErr := job.metaman.Update(job.TaskId, func(md *Metadata) {
			md.Success = false
			md.Message = jobErr.Error()
			md.EndTime = time.Now().UTC()
			md.Output = output.Bytes()
			md.Errput = errput.Bytes()
			md.CalculateTotalSize()
		})
		if mdErr != nil {
			jobErr = fmt.Errorf("cannot save metadata %s. %s.", mdErr, jobErr)
		}
		return jobErr
	}

	return job.metaman.Update(job.TaskId, func(md *Metadata) {
		md.Success = true
		md.Message = "OK"
		md.EndTime = time.Now().UTC()
		md.Output = output.Bytes()
		md.Errput = errput.Bytes()
		md.CalculateTotalSize()
	})
}
