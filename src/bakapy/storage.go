package bakapy

import (
	"fmt"
	"github.com/op/go-logging"
	"net"
	"os"
	"path"
	"path/filepath"
	"time"
)

type StorageCurrentJob struct {
	TaskId      TaskId
	FileAddChan chan JobMetadataFile
	Namespace   string
	Gzip        bool
}

type Storage struct {
	RootDir     string
	MetadataDir string
	currentJobs map[TaskId]StorageCurrentJob
	listenAddr  string
	jobManager  *StorageJobManager
	connections chan *StorageConn
	logger      *logging.Logger
}

func NewStorage(cfg *Config) *Storage {
	return &Storage{
		MetadataDir: cfg.MetadataDir,
		RootDir:     cfg.StorageDir,
		currentJobs: make(map[TaskId]StorageCurrentJob),
		connections: make(chan *StorageConn),
		jobManager:  NewStorageJobManager(),
		listenAddr:  cfg.Listen,
		logger:      logging.MustGetLogger("bakapy.storage"),
	}
}

func (stor *Storage) Start() {
	ln := stor.Listen()
	go stor.Serve(ln)
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

func (stor *Storage) Listen() net.Listener {
	stor.logger.Info("Listening on %s", stor.listenAddr)
	ln, err := net.Listen("tcp", stor.listenAddr)
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				stor.logger.Error("Error during accept() call: %v", err)
				return
			}
			stor.logger.Debug("new connection from %s", conn.RemoteAddr().String())

			loggerName := fmt.Sprintf("bakapy.storage.conn[%s]", conn.RemoteAddr().String())
			logger := logging.MustGetLogger(loggerName)
			stor.connections <- NewStorageConn(stor, conn, logger)
		}

	}()
	return ln
}

func (stor *Storage) GetCurrentJobIds() []TaskId {
	jobs := stor.jobManager.GetJobs()
	keys := make([]TaskId, len(jobs))
	for k := range stor.currentJobs {
		keys = append(keys, k)
	}
	return keys
}

func (stor *Storage) AddJob(currentJob StorageCurrentJob) {
	stor.jobManager.AddJob <- currentJob
}

func (stor *Storage) RemoveJob(id TaskId) {
	stor.jobManager.RemoveJob <- id
}

func (stor *Storage) Serve(ln net.Listener) {
	for {
		go stor.handleConnection(<-stor.connections)
	}
}

func (stor *Storage) handleConnection(conn *StorageConn) {
	var err error
	defer conn.logger.Debug("connection handled")

	if err = conn.ReadTaskId(); err != nil {
		conn.logger.Warning("cannot read task id: %s. closing connection", err)
		return
	}

	stor.jobManager.AddConnection <- conn.TaskId
	defer func() {
		stor.jobManager.RemoveConnection <- conn.TaskId
	}()

	if err = conn.ReadFilename(); err != nil {
		conn.logger.Warning("cannot read filename: %s. closing connection", err)
		return
	}

	if conn.CurrentFilename == JOB_FINISH {
		conn.logger.Warning("got deprecated magic word '%s' as filename, ignoring", JOB_FINISH)
		return
	}

	if err = conn.SaveFile(); err != nil {
		conn.logger.Warning("cannot save file: %s. closing connection", err)
		return
	}
	return
}

func (stor *Storage) GetActiveJob(taskId TaskId) *StorageCurrentJob {
	return stor.jobManager.GetJob(taskId)
}

func (stor *Storage) WaitJob(taskId TaskId) {
	for {
		if stor.jobManager.GetJob(taskId) == nil && !stor.jobManager.HasConnections(taskId) {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}
