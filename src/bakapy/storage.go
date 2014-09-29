package bakapy

import (
	"fmt"
	"github.com/op/go-logging"
	"net"
	"time"
)

type StorageNewJobEvent struct {
	TaskId      TaskId
	FileAddChan chan JobMetadataFile
	Namespace   string
}

type Storage struct {
	RootDir     string
	CurrentJobs map[TaskId]StorageNewJobEvent
	listenAddr  string
	JobsChan    chan StorageNewJobEvent
	connections chan *StorageConn
	logger      *logging.Logger
}

func NewStorage(cfg *Config) *Storage {
	return &Storage{
		RootDir:     cfg.StorageDir,
		CurrentJobs: make(map[TaskId]StorageNewJobEvent),
		JobsChan:    make(chan StorageNewJobEvent, 5),
		connections: make(chan *StorageConn),
		listenAddr:  cfg.Listen,
		logger:      logging.MustGetLogger("bakapy.storage"),
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
	keys := make([]TaskId, 0, len(stor.CurrentJobs))
	for k := range stor.CurrentJobs {
		keys = append(keys, k)
	}
	return keys
}

func (stor *Storage) Serve(ln net.Listener) {
	for {
		select {
		case event := <-stor.JobsChan:
			stor.logger.Info("new job %s", event.TaskId)
			stor.CurrentJobs[event.TaskId] = event
		case conn := <-stor.connections:
			go stor.handleConnection(conn)
		}

	}
}

func (stor *Storage) handleConnection(conn *StorageConn) {
	var err error
	defer conn.logger.Debug("closing connection")
	defer conn.Close()
	if err = conn.ReadTaskId(); err != nil {
		conn.logger.Warning("cannot read task id: %s. closing connection", err)
		return
	}

	if err = conn.ReadFilename(); err != nil {
		conn.logger.Warning("cannot read filename: %s. closing connection", err)
		return
	}

	if conn.CurrentFilename == JOB_FINISH {
		conn.logger.Debug("got magic word '%s' as filename - job finished", JOB_FINISH)
		conn.logger.Info("removing from active jobs list")
		delete(stor.CurrentJobs, conn.TaskId)
		return
	}

	if err = conn.SaveFile(); err != nil {
		conn.logger.Warning("cannot save file: %s. closing connection", err)
		return
	}
}

func (stor *Storage) Wait() {
	for len(stor.CurrentJobs) == 0 {
		time.Sleep(time.Second * 2)
	}
}
