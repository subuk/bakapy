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
	"time"
)

type Storage struct {
	RootDir     string
	metaman     MetaManager
	listenAddr  string
	connections chan *StorageConn
	logger      *logging.Logger
}

func NewStorage(cfg *Config, metaman MetaManager) *Storage {
	return &Storage{
		RootDir:     cfg.StorageDir,
		connections: make(chan *StorageConn),
		listenAddr:  cfg.Listen,
		metaman:     metaman,
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
		go func() {
			err := stor.HandleConnection(NewStorageConn(conn, logger))
			if err != nil {
				stor.logger.Warning("Error during connection from %s: %s", conn.RemoteAddr(), err)
			} else {
				stor.logger.Info("connection from %s handled successfully", conn.RemoteAddr())
			}

		}()
	}
}

func (stor *Storage) HandleConnection(conn StorageProtocolHandler) error {
	var err error

	taskId, err := conn.ReadTaskId()
	if err != nil {
		msg := fmt.Sprintf("cannot read task id: %s. closing connection", err)
		return errors.New(msg)
	}

	metadata, err := stor.metaman.View(taskId)
	if err != nil {
		msg := fmt.Sprintf("Cannot find task id '%s' in current job list, closing connection", taskId)
		return errors.New(msg)
	}

	filename, err := conn.ReadFilename()
	if err != nil {
		msg := fmt.Sprintf("cannot read filename: %s. closing connection", err)
		return errors.New(msg)
	}

	if filename == JOB_FINISH {
		stor.logger.Warning("got deprecated magic word '%s' as filename, ignoring", JOB_FINISH)
		return nil
	}

	fileSavePath := path.Join(
		stor.RootDir,
		metadata.Namespace,
		filename,
	)

	if metadata.Gzip {
		fileSavePath += ".gz"
	}

	fileMeta := MetadataFileEntry{}
	fileMeta.Name = filename
	fileMeta.SourceAddr = conn.RemoteAddr().String()
	fileMeta.StartTime = time.Now()

	stor.logger.Info("saving file %s", fileSavePath)
	err = os.MkdirAll(path.Dir(fileSavePath), 0750)
	if err != nil {
		msg := fmt.Sprintf("cannot create file folder: %s", err)
		return errors.New(msg)
	}

	fd, err := os.Create(fileSavePath)
	if err != nil {
		msg := fmt.Sprintf("cannot open file: %s", err)
		return errors.New(msg)
	}

	var file io.WriteCloser
	var gzWriter io.WriteCloser
	if metadata.Gzip {
		gzWriter = gzip.NewWriter(fd)
		file = gzWriter
	} else {
		file = fd
	}

	stream := bufio.NewWriter(file)
	written, err := conn.ReadContent(stream)
	if err != nil {
		msg := fmt.Sprintf("cannot save file: %s. closing connection", err)
		return errors.New(msg)
	}

	stream.Flush()
	if metadata.Gzip {
		gzWriter.Close()
	}
	fd.Close()

	stor.logger.Debug("adding file %s to metadata", fileMeta.Name)
	fileMeta.Size = written
	fileMeta.EndTime = time.Now()

	err = stor.metaman.Update(taskId, func(md *Metadata) {
		md.Files = append(md.Files, fileMeta)
	})
	if err != nil {
		stor.logger.Critical("cannot save metadata: %s", err.Error())
	}
	return nil
}
