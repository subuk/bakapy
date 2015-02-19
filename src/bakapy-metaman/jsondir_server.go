package main

import (
	"bakapy"
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net"
	"net/rpc"
)

type JSONDirServer struct {
	metaman       *JSONDir
	logger        *logging.Logger
	pendingUpdate map[string]bakapy.TaskId
	secret        string
	listen        string
}

type RPCUpdateArg struct {
	ConnId   *string
	TaskId   *bakapy.TaskId
	Metadata *bakapy.Metadata
	FileMeta *bakapy.MetadataFileEntry
}

func NewJSONDirServer(listen, secret, root string) *JSONDirServer {
	s := &JSONDirServer{
		logger:        logging.MustGetLogger("bakapy.metaman_rpc_server"),
		metaman:       NewJSONDir(root),
		pendingUpdate: make(map[string]bakapy.TaskId),
		listen:        listen,
		secret:        secret,
	}
	return s
}

func (mms *JSONDirServer) Keys(noargs *bool, reply *[]bakapy.TaskId) error {
	keys := []bakapy.TaskId{}
	for key := range mms.metaman.Keys() {
		keys = append(keys, key)
	}
	*reply = keys
	return nil
}

func (mms *JSONDirServer) View(id bakapy.TaskId, response *bakapy.Metadata) error {
	mms.logger.Debug("View: called")
	defer mms.logger.Debug("View: return")
	md, err := mms.metaman.View(id)
	if err != nil {
		return err
	}
	*response = md
	return nil
}

func (mms *JSONDirServer) Add(md *bakapy.Metadata, noreply *bool) error {
	return mms.metaman.Add(md.TaskId, *md)
}

func (mms *JSONDirServer) GetForUpdate(args *RPCUpdateArg, reply *bakapy.Metadata) error {
	mms.logger.Debug("GetForUpdate: called")
	defer mms.logger.Debug("GetForUpdate: return")

	if args.TaskId == nil {
		return fmt.Errorf("args.TaskId is nil")
	}

	if args.ConnId == nil {
		return fmt.Errorf("args.connId is nil")
	}

	taskId := *args.TaskId
	connId := *args.ConnId

	if taskId.String() == "" {
		return fmt.Errorf("args.TaskId must not be blank")
	}

	if connId == "" {
		return fmt.Errorf("args.ConnId must not be blank")
	}

	if _, exist := mms.pendingUpdate[connId]; exist {
		return fmt.Errorf("this client already take metadata for task id %s for update", taskId)
	}

	md, err := mms.metaman.GetForUpdate(taskId)
	if err != nil {
		return err
	}
	mms.logger.Debug("adding taskid %s as %s pending update", taskId, connId)
	mms.pendingUpdate[connId] = taskId
	*reply = *md
	return nil
}

func (mms *JSONDirServer) Save(args *RPCUpdateArg, noreply *bool) error {
	mms.logger.Debug("Save: called")
	defer mms.logger.Debug("Save: return")
	if taskId, exist := mms.pendingUpdate[*args.ConnId]; !exist {
		return fmt.Errorf("%s not locked by this connection", taskId)
	}

	delete(mms.pendingUpdate, *args.ConnId)
	return mms.metaman.Save(args.Metadata.TaskId, args.Metadata)
}

func (mms *JSONDirServer) AddFile(args *RPCUpdateArg, noreply *bool) error {
	if args.FileMeta == nil {
		return fmt.Errorf("args.FileMeta required")
	}
	if args.TaskId == nil {
		return fmt.Errorf("args.TaskId required")
	}
	fm := *args.FileMeta
	id := *args.TaskId
	md, err := mms.metaman.GetForUpdate(id)
	if err != nil {
		return err
	}
	md.Files = append(md.Files, fm)
	return mms.metaman.Save(id, md)
}

func (mms *JSONDirServer) Remove(id bakapy.TaskId, noreply *bool) error {
	mms.logger.Debug("Remove: called")
	defer mms.logger.Debug("Remove: return")
	return mms.metaman.Remove(id)
}

func (mms *JSONDirServer) CleanupConn(connId string) {
	if taskId, exist := mms.pendingUpdate[connId]; exist {
		mms.logger.Debug("cleaning up locks for connection %s", connId)
		mms.metaman.CancelUpdate(taskId)
		return
	}
	mms.logger.Debug("no garbage for connection %s found", connId)
}

func (mms *JSONDirServer) Serve() {

	if err := rpc.RegisterName("Metaman", mms); err != nil {
		panic(err)
	}

	ln, err := net.Listen("tcp", mms.listen)
	if err != nil {
		panic(fmt.Errorf("cannot bind metadata rpc server: %s", err))
	}

	expectedSecret := bakapy.SHA256String(mms.secret)
	for {
		conn, err := ln.Accept()
		if err != nil {
			mms.logger.Warning("error during accept() call: %s", err)
			return
		}
		mms.logger.Debug("new RPC connection from %s", conn.RemoteAddr().String())
		authRequest := make([]byte, 64)
		_, err = io.ReadFull(conn, authRequest)
		if err != nil {
			mms.logger.Warning("failed to read auth info from client %s", conn.RemoteAddr().String())
			conn.Close()
			continue
		}

		if string(authRequest) != expectedSecret {
			mms.logger.Warning("failed to authenticate client %s, bad secret %s", conn.RemoteAddr().String(), authRequest)
			io.WriteString(conn, "00000000-0000-0000-0000-000000000000")
			conn.Close()
			continue
		}
		mms.logger.Info("authentication successfull for %s", conn.RemoteAddr().String())
		connId := uuid.New()
		_, err = io.WriteString(conn, connId)
		if err != nil {
			mms.logger.Warning("cannot send successfull authentication message to client: %s", err)
			conn.Close()
			continue
		}
		go func(connId string) {
			mms.logger.Debug("serving connection for client %s", conn.RemoteAddr().String())
			rpc.ServeConn(conn)
			mms.CleanupConn(connId)
			mms.logger.Debug("connection for client %s closed", conn.RemoteAddr().String())
		}(connId)

	}
}
