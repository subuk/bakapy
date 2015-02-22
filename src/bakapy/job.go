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
	notify      Notificator
}

func NewJob(
	name string,
	cfg *JobConfig,
	StorageAddr string,
	scripts BackupScriptPool,
	executor Executer,
	metaman MetaManager,
	notify Notificator,
) *Job {

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
		notify:      notify,
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

func (job *Job) Run() {
	job.logger.Info("starting up")
	now := time.Now().UTC()
	md := Metadata{
		TaskId:     job.TaskId,
		JobName:    job.Name,
		Gzip:       job.cfg.Gzip,
		Namespace:  job.cfg.Namespace,
		Command:    job.cfg.Command,
		StartTime:  now,
		ExpireTime: now.Add(job.cfg.MaxAge),
	}
	err := job.metaman.Add(job.TaskId, md)
	if err != nil {
		err = fmt.Errorf("cannot add metadata: %s", err)
		job.logger.Critical(err.Error())
		job.notify.MetadataAccessFailed(err)
		return
	}

	script, err := job.getScript()
	if err != nil {
		err = fmt.Errorf("cannot get job script: %s", err)
		job.logger.Critical(err.Error())
		updErr := job.metaman.Update(job.TaskId, func(md *Metadata) {
			md.Message = err.Error()
		})
		if updErr != nil {
			updErr = fmt.Errorf("cannot set metadata error: %s", updErr.Error())
			job.logger.Critical(updErr.Error())
			job.notify.MetadataAccessFailed(updErr)
		} else {
			md.Message = err.Error()
			job.notify.JobFinished(md)
		}
		return
	}

	output := new(bytes.Buffer)
	errput := new(bytes.Buffer)
	job.logger.Debug("execution started")
	jobErr := job.executor.Execute(script, output, errput)
	job.logger.Debug("execution finished: %s", jobErr)
	if jobErr != nil {
		job.logger.Warning("job failed: %s", jobErr.Error())
	}

	job.logger.Debug("Command output: %s", output.String())
	job.logger.Debug("Command errput: %s", errput.String())
	err = job.metaman.Update(job.TaskId, func(mtd *Metadata) {
		job.logger.Debug("Updating task metadata %s", mtd)
		mtd.EndTime = time.Now().UTC()
		mtd.Output = output.Bytes()
		mtd.Errput = errput.Bytes()
		mtd.CalculateTotalSize()

		if jobErr != nil {
			mtd.Success = false
			mtd.Message = jobErr.Error()
		} else {
			mtd.Success = true
			mtd.Message = "OK"
		}
		job.logger.Debug("Updating task metadata done %s", mtd)
	})

	if err != nil {
		err = fmt.Errorf("cannot update metadata: %s", err)
		job.logger.Critical(err.Error())
		job.notify.MetadataAccessFailed(err)
		return
	}

	md, err = job.metaman.View(job.TaskId)
	if err != nil {
		err = fmt.Errorf("cannot get metadata for notification: %s", err)
		job.logger.Warning(err.Error())
		job.notify.MetadataAccessFailed(err)
		return
	}
	job.logger.Debug("notification metadata: %s", md)
	if err := job.notify.JobFinished(md); err != nil {
		job.logger.Warning("failed to run notificator %s: %s", job.notify.Name(), err)
	}
}
