package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Secret       string
	Listen       string
	MetadataAddr string `yaml:"metadata_addr"`
	Root         string
}

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
