package bakapy

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net"
	"os"
	"path"
	"strconv"
	"time"
)

type StorageConnState uint8

type StorageConn struct {
	net.Conn

	CurrentFilename string
	BytesReaded     int
	TaskId          TaskId
	jobEvent        StorageNewJobEvent
	stor            *Storage
	logger          *logging.Logger
	state           StorageConnState
}

func NewStorageConn(stor *Storage, conn net.Conn, logger *logging.Logger) *StorageConn {

	return &StorageConn{
		Conn:   conn,
		stor:   stor,
		logger: logger,
		state:  STATE_WAIT_TASK_ID,
	}
}

func (conn *StorageConn) ReadTaskId() error {
	if conn.state != STATE_WAIT_TASK_ID {
		msg := fmt.Sprintf("protocol error - cannot read task id in state %d", conn.state)
		return errors.New(msg)
	}

	taskId := make([]byte, STORAGE_TASK_ID_LEN)

   conn.logger.Debug("reading task id")
	readed, err := io.ReadFull(conn, taskId)
	conn.logger.Debug("readed %d bytes", readed)

	if err != nil {
		if err == io.EOF {
			return errors.New("received EOF on authentication")
		} else {
			msg := fmt.Sprintf("read error during authentication: %s", err)
			return errors.New(msg)
		}

	}
	conn.TaskId = TaskId(taskId)
	jobEvent, exist := conn.stor.CurrentJobs[conn.TaskId]
	if !exist {
		msg := fmt.Sprintf("Cannot find task id '%s' in current job list (%s), closing connection", taskId, conn.stor.GetCurrentJobIds())
		return errors.New(msg)
	}

	conn.logger.Debug("task id '%s' successfully readed.", taskId)
	conn.jobEvent = jobEvent
	conn.state = STATE_WAIT_FILENAME

	loggerName := fmt.Sprintf("bakapy.storage.conn[%s][%s]", conn.RemoteAddr().String(), conn.TaskId)
	conn.logger = logging.MustGetLogger(loggerName)

	return nil
}

func (conn *StorageConn) ReadFilename() error {
	if conn.state != STATE_WAIT_FILENAME {
		msg := fmt.Sprintf("protocol error - cannot read filename in state %d", conn.state)
		return errors.New(msg)
	}
	conn.logger.Debug("reading filename length")
	var rawFilenameLen = make([]byte, STORAGE_FILENAME_LEN_LEN)
	readed, err := io.ReadFull(conn, rawFilenameLen)
	if err != nil {
		return err
	}
	conn.logger.Debug("readed %d bytes: %s", readed, rawFilenameLen)

	filenameLen, err := strconv.ParseInt(string(rawFilenameLen), 10, 64)
	if err != nil {
		return err
	}

	var filename = make([]byte, filenameLen)
	readed, err = io.ReadFull(conn, filename)
	if err != nil {
		return err
	}
	conn.logger.Debug("readed %d bytes: %s", readed, filename)

	conn.CurrentFilename = string(filename)
	conn.state = STATE_WAIT_DATA
	return nil
}

func (conn *StorageConn) SaveFile() error {
	if conn.state != STATE_WAIT_DATA {
		msg := fmt.Sprintf("protocol error - cannot read data in state %d", conn.state)
		return errors.New(msg)
	}
	savePath := path.Join(
		conn.stor.RootDir,
		conn.jobEvent.Namespace,
		conn.CurrentFilename,
	)
	conn.logger.Info("saving file %s", savePath)
	err := os.MkdirAll(path.Dir(savePath), 0750)
	if err != nil {
		return err
	}

	file, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileMeta := JobMetadataFile{
		Name:       conn.CurrentFilename,
		SourceAddr: conn.RemoteAddr().String(),
		StartTime:  time.Now(),
	}

	conn.state = STATE_RECEIVING

	written, err := io.Copy(file, conn)
	if err != nil {
		return err
	}

	fileMeta.Size = written
	fileMeta.EndTime = time.Now()
	conn.jobEvent.FileAddChan <- fileMeta

	err = file.Sync()
	if err != nil {
		return err
	}
	conn.state = STATE_END
	return nil
}
