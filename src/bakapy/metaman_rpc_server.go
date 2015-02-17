package bakapy

import (
	"github.com/op/go-logging"
)

type MetaRPCServer struct {
	metaman *MetaMan
	logger  *logging.Logger
}

func NewMetaRPCServer(mm *MetaMan) *MetaRPCServer {
	s := &MetaRPCServer{
		logger:  logging.MustGetLogger("bakapy.metaman_rpc_server"),
		metaman: mm,
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

func (mms *MetaRPCServer) GetForUpdate(id TaskId, reply *Metadata) error {
	mms.logger.Debug("GetForUpdate: called")
	defer mms.logger.Debug("GetForUpdate: return")
	md, err := mms.metaman.GetForUpdate(id)
	if err != nil {
		return err
	}
	metadataCopy := *md
	*reply = metadataCopy
	return nil
}

func (mms *MetaRPCServer) Save(md *Metadata, noreply *bool) error {
	mms.logger.Debug("Save: called")
	defer mms.logger.Debug("Save: return")
	return mms.metaman.Save(md.TaskId, md)
}

func (mms *MetaRPCServer) Remove(id TaskId, noreply *bool) error {
	mms.logger.Debug("Remove: called")
	defer mms.logger.Debug("Remove: return")
	return mms.metaman.Remove(id)
}
