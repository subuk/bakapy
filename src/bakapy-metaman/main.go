package main

import (
	"bakapy"
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Config struct {
	Listen string
	Root   string
	Secret string
}

var logger = logging.MustGetLogger("bakapy.metaman")
var CONFIG_PATH = flag.String("config", "metaman.conf", "Path to config file")
var LOG_LEVEL = flag.String("loglevel", "debug", "Log level")
var TEST_CONFIG_ONLY = flag.Bool("test", false, "Check config and exit")

func ParseConfig(configPath string) (*Config, error) {
	config := &Config{}

	rawConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(rawConfig, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

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

	if *TEST_CONFIG_ONLY {
		return
	}
	logger.Info("starting up with configuration %s", config)
	metaServer := NewJSONDirServer(config.Listen, config.Secret, config.Root)
	metaServer.Serve()
}
