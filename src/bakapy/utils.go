package bakapy

import (
	"github.com/op/go-logging"
	"os"
	"path"
	"strings"
)

func SetupLogging(logLevel string) error {
	format := "%{color}%{time:15:04:05} %{level:.8s} %{module} %{message}%{color:reset}"
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
	level, err := logging.LogLevel(strings.ToUpper(logLevel))
	if err != nil {
		return err
	}
	logging.SetLevel(level, "")
	return nil
}

func RunJob(job *Job, config *Config) string {
	logger := logging.MustGetLogger("bakapy.job")
	metadata := job.Run()
	saveTo := path.Join(config.MetadataDir, string(metadata.TaskId))
	err := metadata.Save(saveTo)
	if err != nil {
		logger.Critical("cannot save metadata: %s", err)
	}
	logger.Info("metadata for job %s successfully saved to %s", metadata.TaskId, saveTo)
	if !metadata.Success {
		logger.Critical("job '%s' failed", job.Name)
	} else {
		logger.Info("job '%s' finished", job.Name)
	}
	return saveTo
}
