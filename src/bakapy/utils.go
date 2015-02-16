package bakapy

import (
	"crypto/sha256"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"log/syslog"
	"net"
	"net/rpc"
	"os"
	"strings"
)

func SetupLogging(logLevel string) error {
	format := "%{level:.8s} %{module} %{message}"
	stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)
	syslogBackend, err := logging.NewSyslogBackendPriority("", syslog.LOG_CRIT|syslog.LOG_DAEMON)
	if err != nil {
		return err
	}

	logging.SetBackend(stderrBackend, syslogBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))
	level, err := logging.LogLevel(strings.ToUpper(logLevel))
	if err != nil {
		return err
	}
	logging.SetLevel(level, "")
	return nil
}

func Sha256String(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func ServeRPC(listen, secret string, exports ...interface{}) {
	logger := logging.MustGetLogger("bakapy.rpc")
	for _, rcvr := range exports {
		if err := rpc.Register(rcvr); err != nil {
			panic(err)
		}
	}
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		panic(fmt.Errorf("cannot bind metadata rpc server: %s", err))
	}

	expectedSecret := Sha256String(secret)
	for {
		if conn, err := ln.Accept(); err != nil {
			logger.Warning("error during accept() call: %s", err)
		} else {
			logger.Debug("new RPC connection from %s", conn.RemoteAddr().String())
			authRequest := make([]byte, 64)
			_, err := io.ReadFull(conn, authRequest)
			if err != nil {
				logger.Warning("failed to read auth info from client %s", conn.RemoteAddr().String())
				conn.Close()
				continue
			}

			if string(authRequest) != expectedSecret {
				logger.Warning("failed to authenticate client %s, bad secret %s", conn.RemoteAddr().String(), authRequest)
				io.WriteString(conn, "NOO")
				conn.Close()
				continue
			}
			logger.Info("authentication successfull for %s", conn.RemoteAddr().String())
			_, err = io.WriteString(conn, "YES")
			if err != nil {
				logger.Warning("cannot send successfull authentication message to client: %s", err)
				conn.Close()
				continue
			}
			go rpc.ServeConn(conn)
		}
	}
}
