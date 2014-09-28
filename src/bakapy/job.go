package bakapy

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net"
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

type JobMetadataFile struct {
	Name       string
	Size       int64
	SourceAddr net.Addr
	StartTime  time.Time
	EndTime    time.Time
}

func (m *JobMetadataFile) String() string {
	return fmt.Sprintf(`{name: "%s", size: "%d", start_time: "%s", end_time: "%s"`,
		m.Name, m.Size, m.StartTime, m.EndTime)
}

type JobMetadata struct {
	TaskId     TaskId
	Success    bool
	Message    string
	TotalSize  int64
	StartTime  time.Time
	EndTime    time.Time
	ExpireTime time.Time
	Files      []JobMetadataFile
	Pid        int
	RetCode    uint
	Script     []byte
	Output     []byte
	Errput     []byte
	Config     JobConfig
}

func (metadata *JobMetadata) Save(saveTo string) error {
	err := os.MkdirAll(path.Dir(saveTo), 0750)
	if err != nil {
		return err
	}
	file, err := os.Create(saveTo)
	if err != nil {
		return err
	}
	defer file.Close()
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	return nil
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
	if job.cfg.Sudo {
		remoteCmd = "sudo /bin/bash"
	} else {
		remoteCmd = "/bin/bash"
	}
	args := []string{
		job.gcfg.SSHBin, job.cfg.Host,
		"-oBatchMode=yes",
		"-p", strconv.FormatInt(int64(job.cfg.Port), 10),
		remoteCmd,
	}
	job.logger.Debug(string(script))
	job.logger.Debug("executing command '%s'", strings.Join(args, " "))
	cmd := exec.Cmd{
		Path: args[0],
		Args: args,
	}

	output = new(bytes.Buffer)
	errput = new(bytes.Buffer)

	cmd.Stderr = output
	cmd.Stdout = errput
	cmd.Stdin = bytes.NewReader(script)
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
		Pid:       os.Getpid(),
		Config:    job.cfg,
		StartTime: time.Now(),
		TaskId:    TaskId(uuid.NewUUID().String()),
		Success:   false,
	}
	loggerName := fmt.Sprintf("bakapy.job[%s][%s]", job.Name, metadata.TaskId)
	job.logger = logging.MustGetLogger(loggerName)
	job.logger.Info("starting up with id %s", metadata.TaskId)

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
