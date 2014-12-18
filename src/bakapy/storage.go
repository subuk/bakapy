package bakapy

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Jober interface {
	AddJob(currentJob *StorageCurrentJob)
	RemoveJob(id TaskId)
	WaitJob(taskId TaskId)
}

type StorageCurrentJob struct {
	TaskId      TaskId
	FileAddChan chan JobMetadataFile
	Namespace   string
	Gzip        bool
}

type Storage struct {
	*StorageJobManager
	RootDir     string
	MetadataDir string
	currentJobs map[TaskId]StorageCurrentJob
	listenAddr  string
	connections chan *StorageConn
	logger      *logging.Logger
}

func NewStorage(cfg *Config) *Storage {
	return &Storage{
		StorageJobManager: NewStorageJobManager(),
		MetadataDir:       cfg.MetadataDir,
		RootDir:           cfg.StorageDir,
		currentJobs:       make(map[TaskId]StorageCurrentJob),
		connections:       make(chan *StorageConn),
		listenAddr:        cfg.Listen,
		logger:            logging.MustGetLogger("bakapy.storage"),
	}
}

func (stor *Storage) Start() {
	ln := stor.Listen()
	go stor.Serve(ln)
}

func (stor *Storage) Listen() net.Listener {
	stor.logger.Info("Listening on %s", stor.listenAddr)
	ln, err := net.Listen("tcp", stor.listenAddr)
	if err != nil {
		panic(err)
	}
	return ln
}

func (stor *Storage) Serve(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			stor.logger.Error("Error during accept() call: %v", err)
			return
		}
		stor.logger.Debug("new connection from %s", conn.RemoteAddr().String())

		loggerName := fmt.Sprintf("bakapy.storage.conn[%s]", conn.RemoteAddr().String())
		logger := logging.MustGetLogger(loggerName)
		go stor.HandleConnection(NewStorageConn(conn, logger))
	}
}

func (stor *Storage) HandleConnection(conn StorageProtocolHandler) {
	var err error
	connLogger := conn.Logger()
	defer connLogger.Debug("connection handled")

	taskId, err := conn.ReadTaskId()
	if err != nil {
		connLogger.Warning("cannot read task id: %s. closing connection", err)
		return
	}
	connLogger = conn.Logger()
	currentJob, exist := stor.GetJob(taskId)
	if !exist {
		connLogger.Warning("Cannot find task id '%s' in current job list, closing connection", taskId)
		return
	}

	stor.AddConnection(taskId)
	defer stor.RemoveConnection(taskId)

	filename, err := conn.ReadFilename()
	if err != nil {
		connLogger.Warning("cannot read filename: %s. closing connection", err)
		return
	}

	if filename == JOB_FINISH {
		connLogger.Warning("got deprecated magic word '%s' as filename, ignoring", JOB_FINISH)
		return
	}

	fileSavePath := path.Join(
		stor.RootDir,
		currentJob.Namespace,
		filename,
	)

	if currentJob.Gzip {
		fileSavePath += ".gz"
	}

	fileMeta := JobMetadataFile{}
	fileMeta.Name = filename
	fileMeta.SourceAddr = conn.RemoteAddr().String()
	fileMeta.StartTime = time.Now()

	connLogger.Info("saving file %s", fileSavePath)
	err = os.MkdirAll(path.Dir(fileSavePath), 0750)
	if err != nil {
		connLogger.Warning("cannot create file folder: %s", err)
		return
	}

	fd, err := os.Create(fileSavePath)
	if err != nil {
		connLogger.Warning("cannot open file: %s", err)
		return
	}

	var file io.WriteCloser
	var gzWriter io.WriteCloser
	if currentJob.Gzip {
		gzWriter = gzip.NewWriter(fd)
		file = gzWriter
	} else {
		file = fd
	}

	stream := bufio.NewWriter(file)
	written, err := conn.ReadContent(stream)
	if err != nil {
		connLogger.Warning("cannot save file: %s. closing connection", err)
		return
	}

	stream.Flush()
	if currentJob.Gzip {
		gzWriter.Close()
	}
	fd.Close()

	connLogger.Debug("sending metadata for file %s to job runner", fileMeta.Name)
	fileMeta.Size = written
	fileMeta.EndTime = time.Now()
	currentJob.FileAddChan <- fileMeta
	return
}

func (stor *Storage) CleanupExpired() error {
	visit := func(metaPath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() {
			return nil
		}

		metadata, err := LoadJobMetadata(metaPath)
		if err != nil {
			stor.logger.Warning("corrupt metadata file %s: %s", metaPath, err.Error())
			return err
		}
		if metadata.ExpireTime.After(time.Now()) {
			return nil
		}

		stor.logger.Info("removing files for expired task %s(%s)",
			metadata.JobName, metadata.TaskId)

		removeErrs := false
		for _, fileMeta := range metadata.Files {
			absPath := path.Join(stor.RootDir, metadata.Namespace, fileMeta.Name)
			stor.logger.Info("removing file %s", absPath)
			_, err := os.Stat(absPath)
			if os.IsNotExist(err) {
				stor.logger.Warning("file %s of job %s does not exist", absPath, metadata.TaskId)
				continue
			}
			err = os.Remove(absPath)
			if err != nil {
				removeErrs = true
				stor.logger.Warning("cannot remove file %s: %s", absPath, err.Error())
			}
		}
		if !removeErrs {
			stor.logger.Info("removing metadata %s", metaPath)
		}
		err = os.Remove(metaPath)
		if err != nil {
			stor.logger.Warning("cannot remove file %s: %s", metaPath, err.Error())
			return err
		}

		return nil
	}

	return filepath.Walk(stor.MetadataDir, visit)
}
