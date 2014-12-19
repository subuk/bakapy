package bakapy

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net"
	"strconv"
)

type StorageConnState uint8

type RemoteReader interface {
	Read(p []byte) (n int, err error)
	RemoteAddr() net.Addr
}

type StorageProtocolHandler interface {
	ReadTaskId() (TaskId, error)
	ReadFilename() (string, error)
	ReadContent(output io.Writer) (int64, error)
	RemoteAddr() net.Addr
	Logger() *logging.Logger
}

type StorageConn struct {
	RemoteReader
	currentJob StorageCurrentJob
	logger     *logging.Logger
	State      StorageConnState
}

func NewStorageConn(rReader RemoteReader, logger *logging.Logger) *StorageConn {
	return &StorageConn{
		RemoteReader: rReader,
		logger:       logger,
		State:        STATE_WAIT_TASK_ID,
	}
}

func (sc *StorageConn) Logger() *logging.Logger {
	return sc.logger
}

func (sc *StorageConn) ReadTaskId() (TaskId, error) {
	if sc.State != STATE_WAIT_TASK_ID {
		msg := fmt.Sprintf("protocol error - cannot read task id in state %d", sc.State)
		return TaskId(""), errors.New(msg)
	}

	sc.logger.Debug("reading task id")
	taskIdBuf := make([]byte, STORAGE_TASK_ID_LEN)
	readed, err := io.ReadFull(sc, taskIdBuf)
	sc.logger.Debug("readed %d bytes", readed)
	if err != nil {
		msg := fmt.Sprintf("received error on reading task id: %s", err)
		return TaskId(""), errors.New(msg)
	}

	taskId := TaskId(taskIdBuf)
	sc.logger.Debug("task id '%s' successfully readed.", taskId)
	sc.State = STATE_WAIT_FILENAME
	loggerName := fmt.Sprintf("bakapy.storage.conn[%s][%s]", sc.RemoteAddr().String(), taskId)
	sc.logger = logging.MustGetLogger(loggerName)

	return taskId, nil
}

func (sc *StorageConn) ReadFilename() (string, error) {
	if sc.State != STATE_WAIT_FILENAME {
		msg := fmt.Sprintf("protocol error - cannot read filename in state %d", sc.State)
		return "", errors.New(msg)
	}

	sc.logger.Debug("reading filename length")
	var rawFilenameLen = make([]byte, STORAGE_FILENAME_LEN_LEN)
	readed, err := io.ReadFull(sc, rawFilenameLen)
	if err != nil {
		msg := fmt.Sprintf("error while reading filename length: %s", err)
		return "", errors.New(msg)
	}
	sc.logger.Debug("readed %d bytes: %s", readed, rawFilenameLen)
	filenameLen, err := strconv.ParseInt(string(rawFilenameLen), 10, 64)
	if err != nil {
		msg := fmt.Sprintf("cannot convert readed filename length to integer:%s: %s", rawFilenameLen, err)
		return "", errors.New(msg)
	}

	sc.logger.Debug("reading filename with length %d", filenameLen)
	var filename = make([]byte, filenameLen)
	readed, err = io.ReadFull(sc, filename)
	if err != nil {
		msg := fmt.Sprintf("cannot read filename: %s", err)
		return "", errors.New(msg)
	}
	sc.logger.Debug("readed %d bytes: %s", readed, filename)

	sc.State = STATE_WAIT_DATA
	return string(filename), nil
}

func (sc *StorageConn) ReadContent(output io.Writer) (int64, error) {
	if sc.State != STATE_WAIT_DATA {
		msg := fmt.Sprintf("protocol error - cannot read data in state %d", sc.State)
		return 0, errors.New(msg)
	}

	sc.State = STATE_RECEIVING

	written, err := io.Copy(output, sc)
	if err != nil {
		msg := fmt.Sprintf("read file content error: %s", err)
		return written, errors.New(msg)
	}

	sc.logger.Info("readed %d bytes", written)
	sc.State = STATE_END
	return written, nil
}
