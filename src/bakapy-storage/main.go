package main

import (
	"bakapy"
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var logger = logging.MustGetLogger("bakapy.storage")
var CONFIG_PATH = flag.String("config", "storage.conf", "Path to config file")
var LOG_LEVEL = flag.String("loglevel", "debug", "Log level")
var TEST_CONFIG_ONLY = flag.Bool("test", false, "Check config and exit")

func main() {
	flag.Parse()
	err := bakapy.SetupLogging(*LOG_LEVEL)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	config, err := ParseConfig(*CONFIG_PATH)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %s\n", err)
		os.Exit(1)
	}

	metaman := bakapy.NewMetaManClient(config.MetadataAddr, config.Secret)
	storage := NewStorage(config.Root, config.Listen, metaman)

	if *TEST_CONFIG_ONLY {
		return
	}

	storage.Start()

	done := make(chan bool)
	shutDownSigs := make(chan os.Signal, 1)
	signal.Notify(shutDownSigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-shutDownSigs
		logger.Warning("Got signal %s, gracefully shutting down with 1 minute timeout", sig)
		done <- storage.Shutdown(60)
	}()
	go func() {
		for {
			err := CleanupExpiredJobs(metaman, storage)
			if err != nil {
				logger.Warning("cleanup failed: %s", err.Error())
			}
			time.Sleep(time.Minute)
		}
	}()
	<-done
}
