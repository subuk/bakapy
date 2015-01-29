package bakapy

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"log/syslog"
	"math/rand"
	"net/smtp"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
)

func SetupLogging(logLevel string) error {
	format := "%{level:.8s} %{module} %{message}"
	stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)
	syslogBackend, err := logging.NewSyslogBackendPriority("", syslog.LOG_CRIT|syslog.LOG_DAEMON)
	if err != nil {
		return err
	}

	logging.SetBackend(stderrBackend, syslogBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
	level, err := logging.LogLevel(strings.ToUpper(logLevel))
	if err != nil {
		return err
	}
	logging.SetLevel(level, "")
	return nil
}

type NotificationTemplateContext struct {
	From    string
	To      string
	Subject string
	JobName string
	Message string
	Output  string
	Errput  string
}

type NotificationSender interface {
	SendMail(addr string, from string, to string, msg []byte) error
	SendFailedJobNotification(meta *JobMetadata) error
}

type mailSender struct {
	cfg    *SMTPConfig
	send   func(string, string, string, []byte) error
	notify func(*JobMetadata) error
}

func NewMailSender(cfg SMTPConfig) NotificationSender {
	return &mailSender{cfg: &cfg}
}

// this SendMail makes all the same like smtp.SendMail, but w/o authentication
func (ms *mailSender) SendMail(addr string, from string, to string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Mail(from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

func (ms *mailSender) SendFailedJobNotification(meta *JobMetadata) error {
	// making input data for email
	if ms.cfg.MailTo == "" {
		curUser, err := user.Current()
		if err != nil {
			curUser = &user.User{"0", "0", "root", "root", "/root"}
		}
		ms.cfg.MailTo = curUser.Name
	}

	if ms.cfg.Host == "" {
		hostname, err := os.Hostname()
		if err != nil {
			ms.cfg.Host = "localhost"
		} else {
			ms.cfg.Host = hostname
		}
	}

	if ms.cfg.MailFrom == "" {
		ms.cfg.MailFrom = fmt.Sprintf("bakapy@%s", ms.cfg.Host)
	}

	if ms.cfg.Port == 0 {
		ms.cfg.Port = 25
	}

	var addr string = fmt.Sprintf("%s:%d", ms.cfg.Host, ms.cfg.Port)
	msg := new(bytes.Buffer)
	err := MAIL_TEMPLATE_JOB_FAILED.Execute(msg, NotificationTemplateContext{
		From:    ms.cfg.MailFrom,
		To:      ms.cfg.MailTo,
		Subject: fmt.Sprintf("[bakapy] job '%s' failed", meta.JobName),
		Message: meta.Message,
		Output:  string(meta.Output),
		Errput:  string(meta.Errput),
	})
	if err != nil {
		return err
	}
	// sending email
	err = ms.send(addr, ms.cfg.MailFrom, ms.cfg.MailTo, msg.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func RunJob(jobName string, jConfig *JobConfig, gConfig *Config, storage *Storage) string {
	logger := logging.MustGetLogger("bakapy.job")
	executor := jConfig.executor
	if executor == nil {
		executor = NewBashExecutor(jConfig.Args, jConfig.Host, jConfig.Port, jConfig.Sudo, gConfig.CommandDir, &jConfig.RemoteFilters)
	}
	job := NewJob(
		jobName, jConfig, gConfig.Listen,
		gConfig.CommandDir, storage, executor,
	)

	job.checkTempDirExistance()
	metadata := job.Run()
	job.removeTempDir()
	saveTo := path.Join(gConfig.MetadataDir, string(metadata.TaskId))
	err := metadata.Save(saveTo)
	if err != nil {
		logger.Critical("cannot save metadata: %s", err)
	}
	logger.Info("metadata for job %s successfully saved to %s", metadata.TaskId, saveTo)
	if !metadata.Success {
		logger.Debug("sending failed job notification to current user")
		sender := NewMailSender(gConfig.SMTP)
		if err := sender.SendFailedJobNotification(metadata); err != nil {
			logger.Critical("cannot send failed job notification: %s", err.Error())
		}

		logger.Critical("job '%s' failed", job.Name)
	} else {
		logger.Info("job '%s' finished", job.Name)
	}
	return saveTo
}

var stuff = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(length int) string {
	raw := make([]rune, length)
	rand.Seed(int64(time.Now().Unix()))
	for i := range raw {
		raw[i] = stuff[rand.Intn(len(stuff))]
	}
	return string(raw)
}

func argsToString(args Filter) string {
	if len(args.Params) < 1 {
		return ""
	}

	result := ""
	for key, value := range args.Params {
		result = result + " ARG_" + strings.ToUpper(key) + "=" + value
	}
	return result
}

// runCmdThroughSSH assuming that host already has authorized_keys
func runCmdThroughSSH(cmd string, host string, port uint) (bytes.Buffer, error) {
	var out bytes.Buffer
	sshCmd := exec.Command("ssh", host, "-oBatchMode=yes", "-p"+strconv.FormatUint(uint64(port), 10), cmd)
	sshCmd.Stdout = &out
	err := sshCmd.Run()
	return out, err
}

func (job *Job) hasRemoteFilters() bool {
	if len(job.cfg.RemoteFilters) > 0 {
		return true
	}
	return false
}

func (job *Job) checkTempDirExistance() error {
	// ...and mkdir it, if it's not exist
	if !job.hasRemoteFilters() {
		return nil
	}

	job.logger.Debug("create template directory for remote filters")
	dir := job.cfg.TempDir
	if job.cfg.TempDir == "" {
		dir = "~/.bakapy_filters"
	}

	out, err := runCmdThroughSSH("test -d "+dir+" && echo 1 || echo 0", job.cfg.Host, job.cfg.Port)
	if err != nil {
		return err
	}

	if strings.TrimSpace(out.String()) == "0" {
		out, err = runCmdThroughSSH("mkdir "+dir+" && echo 1 || echo 0", job.cfg.Host, job.cfg.Port)
		if err != nil {
			return err
		}
		out, err = runCmdThroughSSH("chmod 700 "+dir+" && echo 1 || echo 0", job.cfg.Host, job.cfg.Port)
		if err != nil {
			return err
		}
	}

	if strings.TrimSpace(out.String()) == "0" {
		return errors.New("Can't create template filters directory.")
	}

	return nil
}

func (job *Job) removeTempDir() error {
	if !job.hasRemoteFilters() {
		return nil
	}

	job.logger.Debug("remove template directory w/ remote filters")
	dir := job.cfg.TempDir
	if job.cfg.TempDir == "" {
		dir = "~/.bakapy_filters"
	}

	out, err := runCmdThroughSSH("test -d "+dir+" && echo 1 || echo 0", job.cfg.Host, job.cfg.Port)
	if err != nil {
		return err
	}

	if strings.TrimSpace(out.String()) == "1" {
		out, err = runCmdThroughSSH("rm -rf "+dir+"/filter*.sh && echo 1 || echo 0", job.cfg.Host, job.cfg.Port)
		if err != nil {
			return err
		}
	}

	if strings.TrimSpace(out.String()) == "0" {
		return errors.New("Can't clean template filters directory.")
	}

	return nil
}
