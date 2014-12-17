package bakapy

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/op/go-logging"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

type TaskId string

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
	storage     Jober
	cfg         JobConfig
	logger      *logging.Logger
}

func NewJob(name string, cfg JobConfig, StorageAddr string, commandDir string, jober Jober) *Job {
	taskId := TaskId(uuid.NewUUID().String())
	loggerName := fmt.Sprintf("bakapy.job[%s][%s]", name, taskId)
	return &Job{
		Name:        name,
		TaskId:      taskId,
		StorageAddr: StorageAddr,
		CommandDir:  commandDir,
		cfg:         cfg,
		logger:      logging.MustGetLogger(loggerName),
		storage:     jober,
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

func (job *Job) getCmd() (*exec.Cmd, error) {
	var remoteCmd string
	env := make([]string, len(job.cfg.Args))
	for argName, argValue := range job.cfg.Args {
		arg := fmt.Sprintf("%s='%s'", strings.ToUpper(argName), argValue)
		env = append(env, arg)
	}

	if job.cfg.Port == 0 {
		job.cfg.Port = 22
	}

	if job.cfg.Sudo {
		remoteCmd = fmt.Sprintf("sudo %s /bin/bash", strings.Join(env, " "))
	} else {
		remoteCmd = fmt.Sprintf("%s /bin/bash", strings.Join(env, " "))
	}

	var args []string

	if job.cfg.Host != "" {
		args = []string{
			"ssh", job.cfg.Host,
			"-oBatchMode=yes",
			"-p", strconv.FormatInt(int64(job.cfg.Port), 10),
			remoteCmd,
		}
	} else {
		args = []string{
			"bash", "-c",
			remoteCmd,
		}
	}

	cmdPath, err := exec.LookPath(args[0])
	if err != nil {
		return nil, err
	}
	args[0] = cmdPath

	cmd := &exec.Cmd{
		Path: cmdPath,
		Args: args,
	}
	return cmd, nil
}

func (job *Job) execute(script []byte) (output *bytes.Buffer, errput *bytes.Buffer, err error) {
	output = new(bytes.Buffer)
	errput = new(bytes.Buffer)

	cmd, err := job.getCmd()
	if err != nil {
		return output, errput, err
	}

	cmd.Stderr = errput
	cmd.Stdout = output
	cmd.Stdin = bytes.NewReader(script)

	job.logger.Debug(string(script))
	job.logger.Debug("executing command '%s'",
		strings.Join(cmd.Args, " "))

	err = cmd.Start()
	if err != nil {
		return output, errput, err
	}
	err = cmd.Wait()
	if err != nil {
		return output, errput, err
	}
	return output, errput, nil
}

func (job *Job) Run() *JobMetadata {
	metadata := &JobMetadata{
		JobName:   job.Name,
		Gzip:      job.cfg.Gzip,
		Namespace: job.cfg.Namespace,
		Pid:       os.Getpid(),
		Command:   job.cfg.Command,
		Config:    job.cfg,
		StartTime: time.Now(),
		TaskId:    job.TaskId,
		Success:   false,
	}
	metadata.ExpireTime = metadata.StartTime.Add(job.cfg.MaxAge)
	job.logger.Info("starting up")

	script, err := job.getScript()
	metadata.Script = script
	if err != nil {
		job.logger.Warning("cannot get job script: %s", err.Error())
		metadata.Message = err.Error()
		return metadata
	}

	fileAddChan := make(chan JobMetadataFile, 20)

	job.storage.AddJob(&StorageCurrentJob{
		Gzip:        job.cfg.Gzip,
		TaskId:      metadata.TaskId,
		Namespace:   job.cfg.Namespace,
		FileAddChan: fileAddChan,
	})

	go func() {
		for fileMeta := range fileAddChan {
			job.logger.Debug("adding new file metadata: %s", fileMeta.String())
			metadata.Files = append(metadata.Files, fileMeta)
			metadata.TotalSize += fileMeta.Size
		}
		job.logger.Debug("filemeta updater stopped")
	}()

	output, errput, err := job.execute(metadata.Script)
	job.storage.RemoveJob(metadata.TaskId)

	job.logger.Debug("Command output: %s", output.String())
	job.logger.Debug("Command errput: %s", errput.String())

	metadata.Output = output.Bytes()
	metadata.Errput = errput.Bytes()

	if err != nil {
		job.logger.Warning("command failed: %s", err)
		metadata.Success = false
		metadata.Message = err.Error()
		return metadata
	}

	metadata.Success = true
	metadata.Message = "OK"
	metadata.EndTime = time.Now()

	job.logger.Debug("waiting storage")
	job.storage.WaitJob(metadata.TaskId)
	close(fileAddChan)
	return metadata
}

func (job *Job) IsDisabled() bool {
	return job.cfg.Disabled
}
