package bakapy

import (
	"bufio"
	"compress/gzip"
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

type RemoteReader interface {
	Read(p []byte) (n int, err error)
	RemoteAddr() net.Addr
}

type StorageConn struct {
	conn RemoteReader

	CurrentFilename string
	BytesReaded     int
	TaskId          TaskId
	currentJob      StorageCurrentJob
	stor            *Storage
	logger          *logging.Logger
	state           StorageConnState
}

func NewStorageConn(stor *Storage, conn RemoteReader, logger *logging.Logger) *StorageConn {

	return &StorageConn{
		conn:   conn,
		stor:   stor,
		logger: logger,
		state:  STATE_WAIT_TASK_ID,
	}
}

func (sc *StorageConn) Read(p []byte) (n int, err error) {
	return sc.conn.Read(p)
}

func (sc *StorageConn) ReadTaskId() error {
	if sc.state != STATE_WAIT_TASK_ID {
		msg := fmt.Sprintf("protocol error - cannot read task id in state %d", sc.state)
		return errors.New(msg)
	}

	taskId := make([]byte, STORAGE_TASK_ID_LEN)

	sc.logger.Debug("reading task id")
	readed, err := io.ReadFull(sc, taskId)
	sc.logger.Debug("readed %d bytes", readed)

	if err != nil {
		if err == io.EOF {
			return errors.New("received EOF on authentication")
		} else {
			msg := fmt.Sprintf("read error during authentication: %s", err)
			return errors.New(msg)
		}

	}
	sc.TaskId = TaskId(taskId)
	currentJob, jobExist := sc.stor.GetActiveJob(sc.TaskId)
	if !jobExist {
		msg := fmt.Sprintf("Cannot find task id '%s' in current job list, closing connection", taskId)
		return errors.New(msg)
	}

	sc.logger.Debug("task id '%s' successfully readed.", taskId)
	sc.currentJob = currentJob
	sc.state = STATE_WAIT_FILENAME

	loggerName := fmt.Sprintf("bakapy.storage.conn[%s][%s]", sc.conn.RemoteAddr().String(), sc.TaskId)
	sc.logger = logging.MustGetLogger(loggerName)

	return nil
}

func (sc *StorageConn) ReadFilename() error {
	if sc.state != STATE_WAIT_FILENAME {
		msg := fmt.Sprintf("protocol error - cannot read filename in state %d", sc.state)
		return errors.New(msg)
	}
	sc.logger.Debug("reading filename length")
	var rawFilenameLen = make([]byte, STORAGE_FILENAME_LEN_LEN)
	readed, err := io.ReadFull(sc, rawFilenameLen)
	if err != nil {
		return err
	}
	sc.logger.Debug("readed %d bytes: %s", readed, rawFilenameLen)

	filenameLen, err := strconv.ParseInt(string(rawFilenameLen), 10, 64)
	if err != nil {
		return err
	}

	var filename = make([]byte, filenameLen)
	readed, err = io.ReadFull(sc, filename)
	if err != nil {
		return err
	}
	sc.logger.Debug("readed %d bytes: %s", readed, filename)

	sc.CurrentFilename = string(filename)
	sc.state = STATE_WAIT_DATA
	return nil
}

func (sc *StorageConn) SaveFile() error {
	if sc.state != STATE_WAIT_DATA {
		msg := fmt.Sprintf("protocol error - cannot read data in state %d", sc.state)
		return errors.New(msg)
	}

	if sc.currentJob.Gzip {
		sc.CurrentFilename += ".gz"
	}

	savePath := path.Join(
		sc.stor.RootDir,
		sc.currentJob.Namespace,
		sc.CurrentFilename,
	)

	sc.logger.Info("saving file %s", savePath)
	err := os.MkdirAll(path.Dir(savePath), 0750)
	if err != nil {
		return err
	}

	fd, err := os.Create(savePath)
	if err != nil {
		return err
	}

	var file io.Writer
	var gzWriter *gzip.Writer
	if sc.currentJob.Gzip {
		gzWriter = gzip.NewWriter(fd)
		file = gzWriter
	} else {
		file = fd
	}

	fileMeta := JobMetadataFile{
		Name:       sc.CurrentFilename,
		SourceAddr: sc.conn.RemoteAddr().String(),
		StartTime:  time.Now(),
	}

	sc.state = STATE_RECEIVING

	buff := bufio.NewWriter(file)
	written, err := io.Copy(buff, sc)
	if err != nil {
		return err
	}
	buff.Flush()
	if sc.currentJob.Gzip {
		gzWriter.Close()
	}
	fd.Close()

	fileMeta.Size = written
	fileMeta.EndTime = time.Now()
	sc.logger.Debug("sending metadata for file %s to job runner", fileMeta.Name)
	sc.currentJob.FileAddChan <- fileMeta

	sc.logger.Info("file saved %s", savePath)
	sc.state = STATE_END
	return nil
}
