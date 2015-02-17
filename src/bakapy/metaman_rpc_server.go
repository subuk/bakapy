package bakapy

import (
	"fmt"
	"github.com/op/go-logging"
)

type MetaRPCServer struct {
	metaman       *MetaMan
	logger        *logging.Logger
	pendingUpdate map[string]TaskId
}

type RPCUpdateArg struct {
	ConnId   *string
	TaskId   *TaskId
	Metadata *Metadata
}

func NewMetaRPCServer(mm *MetaMan) *MetaRPCServer {
	s := &MetaRPCServer{
		logger:        logging.MustGetLogger("bakapy.metaman_rpc_server"),
		metaman:       mm,
		pendingUpdate: make(map[string]TaskId),
	}
	return s
}

func (mms *MetaRPCServer) Keys(noargs *bool, reply *[]TaskId) error {
	keys := []TaskId{}
	for key := range mms.metaman.Keys() {
		keys = append(keys, key)
	}
	*reply = keys
	return nil
}

func (mms *MetaRPCServer) View(id TaskId, response *Metadata) error {
	mms.logger.Debug("View: called")
	defer mms.logger.Debug("View: return")
	md, err := mms.metaman.View(id)
	if err != nil {
		return err
	}
	*response = md
	return nil
}

func (mms *MetaRPCServer) Add(md *Metadata, noreply *bool) error {
	return mms.metaman.Add(md.TaskId, *md)
}

func (mms *MetaRPCServer) GetForUpdate(args *RPCUpdateArg, reply *Metadata) error {
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

func (mms *MetaRPCServer) Save(args *RPCUpdateArg, noreply *bool) error {
	mms.logger.Debug("Save: called")
	defer mms.logger.Debug("Save: return")
	if taskId, exist := mms.pendingUpdate[*args.ConnId]; !exist {
		return fmt.Errorf("%s not locked by this connection", taskId)
	}

	delete(mms.pendingUpdate, *args.ConnId)
	return mms.metaman.Save(args.Metadata.TaskId, args.Metadata)
}

func (mms *MetaRPCServer) Remove(id TaskId, noreply *bool) error {
	mms.logger.Debug("Remove: called")
	defer mms.logger.Debug("Remove: return")
	return mms.metaman.Remove(id)
}

func (mms *MetaRPCServer) CleanupConn(connId string) {
	if taskId, exist := mms.pendingUpdate[connId]; exist {
		mms.logger.Debug("cleaning up locks for connection %s", connId)
		mms.metaman.CancelUpdate(taskId)
		return
	}
	mms.logger.Debug("no garbage for connection %s found", connId)
}
