package bakapy

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net"
	"time"
)

type StorageConnState uint8

type StorageConn struct {
	net.Conn

	JobMeta         *JobMetadata
	CurrentFilename string
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

	deadline := time.Now().Add(time.Second * STORAGE_AUTH_TIMEOUT)
	conn.SetDeadline(deadline)
	conn.logger.Debug("reading task id until %s", deadline)

	readed, err := conn.Read(taskId)
	conn.logger.Debug("readed %d bytes", readed)

	if err != nil {
		if err == io.EOF {
			return errors.New("received EOF on authentication")
		} else {
			msg := fmt.Sprintf("read error during authentication: %s", err)
			return errors.New(msg)
		}

	}

	jobMeta, exist := conn.stor.CurrentJobs[TaskId(taskId)]
	if !exist {
		msg := fmt.Sprintf("Cannot find task id '%s' in current job list (%s), closing connection", taskId, conn.stor.GetPlannedJobIds())
		return errors.New(msg)
	}

	conn.logger.Debug("task id '%s' successfully readed.", jobMeta.TaskId)
	conn.JobMeta = jobMeta
	conn.state = STATE_WAIT_FILENAME
	return nil
}

func (conn *StorageConn) ReadFilename() error {
	if conn.state != STATE_WAIT_TASK_ID {
		msg := fmt.Sprintf("protocol error - cannot read filename in state %d", conn.state)
		return errors.New(msg)
	}

	//

	conn.state = STATE_WAIT_DATA
	return nil
}

func (conn *StorageConn) FileSave() error {
	if conn.state != STATE_WAIT_DATA {
		msg := fmt.Sprintf("protocol error - cannot read data in state %d", conn.state)
		return errors.New(msg)
	}
	conn.state = STATE_RECEIVING
	//

	conn.state = STATE_END
	return nil
}
