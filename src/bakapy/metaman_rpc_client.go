package bakapy

import (
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net"
	"net/rpc"
)

type MetaManClient struct {
	client     *rpc.Client
	serverAddr string
	logger     *logging.Logger
	secret     string
}

func NewMetaManClient(serverAddr, secret string) *MetaManClient {
	return &MetaManClient{
		serverAddr: serverAddr,
		logger:     logging.MustGetLogger("bakapy.metaman_rpc_client"),
		secret:     secret,
	}
}

func (mmc *MetaManClient) call(serviceMethod string, args interface{}, reply interface{}) error {
auth:
	if mmc.client == nil {
		conn, err := net.Dial("tcp", mmc.serverAddr)
		if err != nil {
			return fmt.Errorf("cannot connect to %s: %s", mmc.serverAddr, err)
		}

		_, err = io.WriteString(conn, Sha256String(mmc.secret))
		if err != nil {
			return fmt.Errorf("cannot write auth request: %s", err)
		}
		authResponse := make([]byte, 3)
		_, err = io.ReadFull(conn, authResponse)
		if err != nil {
			return fmt.Errorf("cannot read auth response: %s", err)
		}
		if string(authResponse) != "YES" {
			return fmt.Errorf("auth failed: %s", authResponse)
		}

		mmc.client = rpc.NewClient(conn)
		mmc.logger.Debug("rpc connection established")
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
	err := mmc.call("MetaRPCServer.Keys", &args, &reply)
	if err != nil {
		mmc.logger.Warning("error during MetaRPCServer.Keys call: %s", err)
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
	err := mmc.call("MetaRPCServer.View", &id, &md)
	return md, err
}

func (mmc *MetaManClient) Add(id TaskId, md Metadata) error {
	md.TaskId = id
	noreply := false
	return mmc.call("MetaRPCServer.Add", &md, &noreply)
}

func (mmc *MetaManClient) Update(id TaskId, up func(*Metadata)) error {
	noreply := false
	md := &Metadata{}
	err := mmc.call("MetaRPCServer.GetForUpdate", id, md)
	if err != nil {
		return fmt.Errorf("cannot get metadata for update: %s", err)
	}
	up(md)
	return mmc.call("MetaRPCServer.Save", md, &noreply)
}

func (mmc *MetaManClient) Remove(id TaskId) error {
	noreply := false
	return mmc.call("MetaRPCServer.Remove", &id, &noreply)
}
