package bakapy

import (
	"github.com/op/go-logging"
)

type LogWriter struct {
	logger *logging.Logger
	log    func(string)
}

func NewLogWriter(logger *logging.Logger) *LogWriter {
	logfn := func(msg string) {
		logger.Info(msg)
	}

	return &LogWriter{
		logger: logger,
		log:    logfn,
	}
}

func (lw *LogWriter) Write(line []byte) (n int, err error) {
	lw.log(string(line))
	return len(line), nil
}
