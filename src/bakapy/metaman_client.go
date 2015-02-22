package bakapy

import (
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net"
	"net/rpc"
)

type RPCUpdateArg struct {
	ConnId   *string
	TaskId   *TaskId
	Metadata *Metadata
	FileMeta *MetadataFileEntry
}

type MetaManClient struct {
	client     *rpc.Client
	serverAddr string
	logger     *logging.Logger
	secret     string
	connId     string
}

func NewMetaManClient(serverAddr, secret string) *MetaManClient {
	return &MetaManClient{
		serverAddr: serverAddr,
		logger:     logging.MustGetLogger("bakapy.metaman_client"),
		secret:     secret,
	}
}

func (mmc *MetaManClient) call(serviceMethod string, args interface{}, reply interface{}) error {
	mmc.logger.Debug("%s: call", serviceMethod)
	defer mmc.logger.Debug("%s: return", serviceMethod)
auth:
	if mmc.client == nil {
		conn, err := net.Dial("tcp", mmc.serverAddr)
		if err != nil {
			return fmt.Errorf("cannot connect to %s: %s", mmc.serverAddr, err)
		}

		_, err = io.WriteString(conn, SHA256String(mmc.secret))
		if err != nil {
			return fmt.Errorf("cannot write auth request: %s", err)
		}
		authResponse := make([]byte, 36)
		_, err = io.ReadFull(conn, authResponse)
		if err != nil {
			return fmt.Errorf("cannot read auth response: %s", err)
		}
		if string(authResponse) == "00000000-0000-0000-0000-000000000000" {
			return fmt.Errorf("auth failed: %s", authResponse)
		}

		mmc.client = rpc.NewClient(conn)
		mmc.connId = string(authResponse)
		mmc.logger.Debug("rpc connection established with connection id %s", mmc.connId)
	}

	err := mmc.client.Call(serviceMethod, args, reply)
	if err == rpc.ErrShutdown {
		mmc.logger.Debug("connection shutted down")
		mmc.client = nil
		goto auth
	}
	return err
}

func (mmc *MetaManClient) Keys() chan TaskId {
	args := false
	ch := make(chan TaskId)
	var reply []TaskId
	err := mmc.call("Metaman.Keys", &args, &reply)
	if err != nil {
		mmc.logger.Warning("error during Metaman.Keys call: %s", err)
		close(ch)
		return ch
	}
	go func() {
		for _, key := range reply {
			ch <- key
		}
		close(ch)
	}()
	return ch
}

func (mmc *MetaManClient) View(id TaskId) (Metadata, error) {
	var md Metadata
	err := mmc.call("Metaman.View", &id, &md)
	return md, err
}

func (mmc *MetaManClient) Add(id TaskId, md Metadata) error {
	md.TaskId = id
	noreply := false
	return mmc.call("Metaman.Add", &md, &noreply)
}

func (mmc *MetaManClient) AddFile(id TaskId, fm MetadataFileEntry) error {
	noreply := false
	args := &RPCUpdateArg{
		TaskId:   &id,
		FileMeta: &fm,
	}
	return mmc.call("Metaman.AddFile", &args, &noreply)
}

func (mmc *MetaManClient) Update(id TaskId, up func(*Metadata)) error {
	noreply := false
	md := &Metadata{}
	args := &RPCUpdateArg{
		TaskId: &id,
		ConnId: &mmc.connId,
	}
	err := mmc.call("Metaman.GetForUpdate", args, md)
	if err != nil {
		return fmt.Errorf("cannot get metadata for update: %s", err)
	}
	up(md)
	args.Metadata = md
	return mmc.call("Metaman.Save", args, &noreply)
}

func (mmc *MetaManClient) Remove(id TaskId) error {
	noreply := false
	return mmc.call("Metaman.Remove", &id, &noreply)
}
