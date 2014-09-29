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

type Job struct {
	Name     string
	storJobs chan StorageNewJobEvent
	cfg      JobConfig
	gcfg     *Config
	logger   *logging.Logger
}

type TaskId string

type JobTemplateContext struct {
	Meta             *JobMetadata
	GCfg             *Config
	JCfg             JobConfig
	FINISH_MAGIC     string
	FILENAME_LEN_LEN uint
}

func (jctx *JobTemplateContext) ToHost() string {
	return strings.Split(jctx.GCfg.Listen, ":")[0]
}

func (jctx *JobTemplateContext) ToPort() string {
	return strings.Split(jctx.GCfg.Listen, ":")[1]
}

func (job *Job) GetScript(metadata *JobMetadata) ([]byte, error) {
	script := new(bytes.Buffer)
	err := JOB_TEMPLATE.Execute(script, &JobTemplateContext{
		Meta:             metadata,
		GCfg:             job.gcfg,
		JCfg:             job.cfg,
		FINISH_MAGIC:     JOB_FINISH,
		FILENAME_LEN_LEN: STORAGE_FILENAME_LEN_LEN,
	})
	if err != nil {
		return nil, err
	}

	scriptPath := path.Join(job.gcfg.CommandDir, job.cfg.Command)
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

func (job *Job) execute(script []byte) (output *bytes.Buffer, errput *bytes.Buffer, err error) {
	var remoteCmd string
	env := make([]string, len(job.cfg.Args))
	for argName, argValue := range job.cfg.Args {
		arg := fmt.Sprintf("%s='%s'", strings.ToUpper(argName), argValue)
		env = append(env, arg)
	}

	if job.cfg.Sudo {
		remoteCmd = fmt.Sprintf("sudo %s /bin/bash", strings.Join(env, " "))
	} else {
		remoteCmd = fmt.Sprintf("%s /bin/bash", strings.Join(env, " "))
	}

	sshBin, err := exec.LookPath("ssh")
	if err != nil {
		return new(bytes.Buffer), new(bytes.Buffer), err
	}

	args := []string{
		sshBin, job.cfg.Host,
		"-oBatchMode=yes",
		"-p", strconv.FormatInt(int64(job.cfg.Port), 10),
		remoteCmd,
	}

	job.logger.Debug(string(script))
	cmd := exec.Cmd{
		Path: args[0],
		Args: args,
	}

	output = new(bytes.Buffer)
	errput = new(bytes.Buffer)

	cmd.Stderr = errput
	cmd.Stdout = output
	cmd.Stdin = bytes.NewReader(script)

	job.logger.Debug("executing command '%s'",
		strings.Join(args, " "))

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
		Namespace: job.cfg.Namespace,
		Pid:       os.Getpid(),
		Command:   job.cfg.Command,
		Config:    job.cfg,
		StartTime: time.Now(),
		TaskId:    TaskId(uuid.NewUUID().String()),
		Success:   false,
	}
	loggerName := fmt.Sprintf("bakapy.job[%s][%s]", job.Name, metadata.TaskId)
	job.logger = logging.MustGetLogger(loggerName)
	job.logger.Info("starting up")

	script, err := job.GetScript(metadata)
	metadata.Script = script
	if err != nil {
		job.logger.Warning("cannot get job script: %s", err.Error())
		metadata.Message = err.Error()
		return metadata
	}

	fileAddChan := make(chan JobMetadataFile, 5)
	job.storJobs <- StorageNewJobEvent{
		TaskId:      metadata.TaskId,
		Namespace:   job.cfg.Namespace,
		FileAddChan: fileAddChan,
	}

	stopUpdater := make(chan int, 1)
	go func() {
		for {
			select {
			case meta := <-fileAddChan:
				job.logger.Debug("Adding new file metadata: %s", meta.String())
				metadata.Files = append(metadata.Files, meta)
			case <-stopUpdater:
				job.logger.Debug("Stopping metadata update routine")
				return
			}
		}
	}()

	output, errput, err := job.execute(metadata.Script)

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
	metadata.ExpireTime = time.Now().Add(time.Hour * 24 * time.Duration(job.cfg.MaxAgeDays))
	metadata.TotalSize = 0
	for _, fileMeta := range metadata.Files {
		metadata.TotalSize += fileMeta.Size
	}
	stopUpdater <- 1
	return metadata
}

func (job *Job) IsDisabled() bool {
	return job.cfg.Disabled
}

func NewJob(name string, cfg JobConfig, storJobs chan StorageNewJobEvent, globalConfig *Config) *Job {
	loggerName := fmt.Sprintf("bakapy.job[%s][not-started]", name)
	return &Job{
		Name:     name,
		cfg:      cfg,
		logger:   logging.MustGetLogger(loggerName),
		storJobs: storJobs,
		gcfg:     globalConfig,
	}
}
