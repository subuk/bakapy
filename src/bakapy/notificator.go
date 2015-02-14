package bakapy

import (
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Notificator interface {
	JobFinished(md Metadata) error
	Name() string
}

type ScriptedNotificator struct {
	rootDir string
	name    string
	params  map[string]string
	output  io.Writer
	errput  io.Writer
}

func NewScriptedNotificator(rootDir, name string, params map[string]string) *ScriptedNotificator {
	return &ScriptedNotificator{
		rootDir: rootDir,
		params:  params,
		name:    name,
		output:  os.Stdout,
		errput:  os.Stderr,
	}
}

func (s *ScriptedNotificator) JobFinished(md Metadata) error {
	fullPath := path.Join(s.rootDir, "notify-"+s.name+".sh")
	cmd := exec.Command(fullPath)
	cmd.Stdout = s.output
	cmd.Stderr = s.errput

	env := os.Environ()
	if md.Success {
		env = append(env, "BAKAPY_METADATA_SUCCESS=1")
	} else {
		env = append(env, "BAKAPY_METADATA_SUCCESS=0")
	}
	env = append(env, "BAKAPY_METADATA_JOBNAME="+md.JobName)
	env = append(env, "BAKAPY_METADATA_TASKID="+md.TaskId.String())
	env = append(env, "BAKAPY_METADATA_MESSAGE="+md.Message)
	env = append(env, "BAKAPY_METADATA_OUTPUT="+string(md.Output))
	env = append(env, "BAKAPY_METADATA_ERRPUT="+string(md.Errput))
	for key, value := range s.params {
		env = append(env, "BAKAPY_PARAM_"+strings.ToUpper(key)+"="+value)
	}
	cmd.Env = env
	return cmd.Run()
}

func (s *ScriptedNotificator) Name() string {
	return s.name
}
