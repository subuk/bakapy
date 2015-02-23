package bakapy

import (
	"crypto/sha256"
	"fmt"
	"github.com/op/go-logging"
	"log/syslog"
	"os"
	"strings"
)

func SetupLogging(logLevel string) error {
	format := "%{level:.8s} %{module} %{message}"
	stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)
	syslogBackend, err := logging.NewSyslogBackendPriority("", syslog.LOG_INFO|syslog.LOG_DAEMON)
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

func SHA256String(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
