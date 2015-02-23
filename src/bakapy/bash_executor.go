package bakapy

import (
	"bytes"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type BashExecutor struct {
	Args   map[string]string
	Host   string
	Port   uint
	Sudo   bool
	logger *logging.Logger
}

func NewBashExecutor(args map[string]string, host string, port uint, sudo bool) *BashExecutor {
	return &BashExecutor{
		Args:   args,
		Host:   host,
		Port:   port,
		Sudo:   sudo,
		logger: logging.MustGetLogger("bakapy.executor.ssh"),
	}
}

func (e *BashExecutor) GetCmd() (*exec.Cmd, error) {
	var remoteCmd string
	env := make([]string, len(e.Args))
	for argName, argValue := range e.Args {
		arg := fmt.Sprintf("%s='%s'", strings.ToUpper(argName), argValue)
		env = append(env, arg)
	}

	if e.Port == 0 {
		e.Port = 22
	}

	if e.Sudo {
		remoteCmd = fmt.Sprintf("sudo %s /bin/bash", strings.Join(env, " "))
	} else {
		remoteCmd = fmt.Sprintf("%s /bin/bash", strings.Join(env, " "))
	}

	var args []string

	if e.Host != "" {
		args = []string{
			"ssh", e.Host,
			"-oBatchMode=yes",
			"-p", strconv.FormatInt(int64(e.Port), 10),
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

func (e *BashExecutor) Execute(script []byte, output io.Writer, errput io.Writer) error {
	cmd, err := e.GetCmd()
	if err != nil {
		return err
	}

	cmd.Stderr = errput
	cmd.Stdout = output
	cmd.Stdin = bytes.NewReader(script)

	e.logger.Debug(string(script))
	e.logger.Debug("executing command '%s'",
		strings.Join(cmd.Args, " "))

	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
