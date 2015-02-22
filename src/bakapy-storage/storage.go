package main

import (
	"bakapy"
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net"
	"os"
	"path"
	"sync"
	"time"
)

type Storage interface {
	Serve(ln net.Listener)
	Remove(namespace, filename string) error
}

type LocalFileStorage struct {
	RootDir    string
	metaman    bakapy.MetaManager
	listenAddr string
	cons       *sync.WaitGroup
	shutdown   chan int
	logger     *logging.Logger
}

func NewStorage(root, listen string, metaman bakapy.MetaManager) *LocalFileStorage {
	return &LocalFileStorage{
		RootDir:    root,
		listenAddr: listen,
		metaman:    metaman,
		shutdown:   make(chan int),
		cons:       new(sync.WaitGroup),
		logger:     logging.MustGetLogger("bakapy.storage"),
	}
}

func (stor *LocalFileStorage) Start() {
	ln := stor.Listen()
	go stor.Serve(ln)
}

func (stor *LocalFileStorage) Shutdown(seconds int) bool {

	doneCh := make(chan bool)
	timer := time.NewTimer(time.Duration(time.Second * time.Duration(seconds)))

	go func() {
		stor.logger.Debug("gracefully shutdown requested")
		stor.shutdown <- 1
		stor.logger.Debug("wating existing connections to finish")
		stor.cons.Wait()
		doneCh <- true
	}()
	select {
	case <-doneCh:
		return true
	case <-timer.C:
		stor.logger.Warning("gracefully shutdown timed out in %d seconds", seconds)
		return false
	}
}

func (stor *LocalFileStorage) Listen() net.Listener {
	stor.logger.Info("Listening on %s", stor.listenAddr)
	ln, err := net.Listen("tcp", stor.listenAddr)
	if err != nil {
		panic(err)
	}
	return ln
}

type __acc struct {
	c net.Conn
	e error
}

func (stor *LocalFileStorage) Serve(ln net.Listener) {
	acceptCh := make(chan net.Conn)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				stor.logger.Error("Error during accept() call: %v", err)
				return
			}
			acceptCh <- conn
		}
		close(acceptCh)
	}()
	for {
		select {
		case conn := <-acceptCh:
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
				conn.Close()
			}()
		case <-stor.shutdown:
			stor.logger.Debug("Closing listening socket %s", ln.Addr().String())
			if err := ln.Close(); err != nil {
				panic(fmt.Errorf("cannot close listening socket: %s", err))
			}
			return
		}
	}
}

func (stor *LocalFileStorage) Remove(ns, filename string) error {
	fullPath := path.Join(stor.RootDir, ns, filename)
	return os.Remove(fullPath)
}

func (stor *LocalFileStorage) HandleConnection(conn StorageProtocolHandler) error {
	stor.cons.Add(1)
	defer stor.cons.Done()
	var err error

	taskId, err := conn.ReadTaskId()
	if err != nil {
		return fmt.Errorf("cannot read task id: %s. closing connection", err)
	}

	md, err := stor.metaman.View(taskId)
	if err != nil {
		return fmt.Errorf("cannot find task id %s: %s", taskId, err)
	}

	if !md.EndTime.IsZero() {
		return fmt.Errorf("task with id '%s' already finished, closing connection", taskId)
	}

	filename, err := conn.ReadFilename()
	if err != nil {
		return fmt.Errorf("cannot read filename: %s. closing connection", err)
	}

	if filename == bakapy.JOB_FINISH {
		stor.logger.Warning("got deprecated magic word '%s' as filename, ignoring", bakapy.JOB_FINISH)
		return nil
	}

	fileSavePath := path.Join(
		stor.RootDir,
		md.Namespace,
		filename,
	)

	if md.Gzip {
		fileSavePath += ".gz"
	}

	fileMeta := bakapy.MetadataFileEntry{}
	fileMeta.Name = filename
	fileMeta.SourceAddr = conn.RemoteAddr().String()
	fileMeta.StartTime = time.Now()

	stor.logger.Info("saving file %s", fileSavePath)
	err = os.MkdirAll(path.Dir(fileSavePath), 0750)
	if err != nil {
		return fmt.Errorf("cannot create file folder: %s", err)
	}

	fd, err := os.Create(fileSavePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %s", err)
	}

	var file io.WriteCloser
	var gzWriter io.WriteCloser
	if md.Gzip {
		gzWriter = gzip.NewWriter(fd)
		file = gzWriter
	} else {
		file = fd
	}
	stream := bufio.NewWriter(file)
	written, err := conn.ReadContent(stream)
	if err != nil {
		return fmt.Errorf("cannot save file: %s. closing connection", err)
	}

	stream.Flush()
	if md.Gzip {
		gzWriter.Close()
	}
	fd.Close()
	fileMeta.Size = written
	fileMeta.EndTime = time.Now()

	err = stor.metaman.AddFile(taskId, fileMeta)
	if err != nil {
		stor.logger.Critical("cannot save metadata: %s", err.Error())
		return err
	}
	return nil
}