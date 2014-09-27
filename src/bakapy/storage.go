package bakapy

import (
	"fmt"
	"github.com/op/go-logging"
	"net"
)

type StorageTaskChan chan *JobMetadata

type Storage struct {
	RootDir     string
	CurrentJobs map[TaskId]*JobMetadata
	listenAddr  string
	Tasks       StorageTaskChan
	connections chan *StorageConn
	logger      *logging.Logger
}

func NewStorage(cfg *Config) *Storage {
	return &Storage{
		RootDir:     cfg.StorageDir,
		CurrentJobs: make(map[TaskId]*JobMetadata),
		Tasks:       make(StorageTaskChan),
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
			stor.logger.Info("New connection from %s", conn.RemoteAddr().String())

			loggerName := fmt.Sprintf("bakapy.storage.conn[%s]", conn.RemoteAddr().String())
			logger := logging.MustGetLogger(loggerName)
			stor.connections <- NewStorageConn(stor, conn, logger)
		}

	}()
	return ln
}

func (stor *Storage) GetPlannedJobIds() []TaskId {
	keys := make([]TaskId, 0, len(stor.CurrentJobs))
	for k := range stor.CurrentJobs {
		keys = append(keys, k)
	}
	return keys
}

func (stor *Storage) Serve(ln net.Listener) {
	for {
		select {
		case jobMeta := <-stor.Tasks:
			stor.CurrentJobs[jobMeta.TaskId] = jobMeta
		case conn := <-stor.connections:
			go stor.handleConnection(conn)
		}

	}
}

func (stor *Storage) handleConnection(conn *StorageConn) {
	var err error
	defer conn.Close()
	if err = conn.ReadTaskId(); err != nil {
		conn.logger.Warning("cannot read task id: %s. closing connection", err)
		return
	}

	if err = conn.ReadFilename(); err != nil {
		conn.logger.Warning("cannot filename: %s. closing connection.", err)
		return
	}

}
