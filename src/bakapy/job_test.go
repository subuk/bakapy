package bakapy

import (
	"strings"
	"testing"
)

func TestJobGetCmdLocal(t *testing.T) {
	cfg := &JobConfig{
		Args: map[string]string{
			"test": "oneone",
			"two":  "xxx",
		},
	}
	gcfg := &Config{}
	storage := NewStorage(gcfg)
	job := NewJob("test-job", *cfg, gcfg, storage)
	cmd, err := job.GetCmd()
	if err != nil {
		t.Fatal("Error:", err)
	}
	if strings.Join(cmd.Args, "|||") != "/bin/bash|||-c|||  TEST='oneone' TWO='xxx' /bin/bash" {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"))
	}
	t.Log(cmd.Args)
}

func TestJobGetCmdLocalSudo(t *testing.T) {
	cfg := &JobConfig{
		Sudo: true,
		Args: map[string]string{
			"test": "oneone",
			"two":  "xxx",
		},
	}
	gcfg := &Config{}
	storage := NewStorage(gcfg)
	job := NewJob("test-job", *cfg, gcfg, storage)
	cmd, err := job.GetCmd()
	if err != nil {
		t.Fatal("Error:", err)
	}
	if strings.Join(cmd.Args, "|||") != "/bin/bash|||-c|||sudo   TEST='oneone' TWO='xxx' /bin/bash" {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"))
	}
	t.Log(cmd.Args)
}

func TestJobGetCmdRemote(t *testing.T) {
	cfg := &JobConfig{
		Host: "test-host.example",
		Port: 2424,
		Args: map[string]string{
			"test": "oneone",
			"two":  "xxx",
		},
	}
	gcfg := &Config{}
	storage := NewStorage(gcfg)
	job := NewJob("test-job", *cfg, gcfg, storage)
	cmd, err := job.GetCmd()
	if err != nil {
		t.Fatal("Error:", err)
	}
	if strings.Join(cmd.Args, "|||") != "/usr/bin/ssh|||test-host.example|||-oBatchMode=yes|||-p|||2424|||  TEST='oneone' TWO='xxx' /bin/bash" {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"))
	}
	t.Log(cmd.Args)
}

func TestJobGetCmdRemoteSudo(t *testing.T) {
	cfg := &JobConfig{
		Host: "test-host.example",
		Sudo: true,
		Port: 2424,
		Args: map[string]string{
			"test": "oneone",
			"two":  "xxx",
		},
	}
	gcfg := &Config{}
	storage := NewStorage(gcfg)
	job := NewJob("test-job", *cfg, gcfg, storage)
	cmd, err := job.GetCmd()
	if err != nil {
		t.Fatal("Error:", err)
	}
	if strings.Join(cmd.Args, "|||") != "/usr/bin/ssh|||test-host.example|||-oBatchMode=yes|||-p|||2424|||sudo   TEST='oneone' TWO='xxx' /bin/bash" {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"))
	}
	t.Log(cmd.Args)
}

func TestJobGetCmdRemoteNoPort(t *testing.T) {
	cfg := &JobConfig{
		Host: "test-host.example",
		Args: map[string]string{
			"test": "oneone",
			"two":  "xxx",
		},
	}
	gcfg := &Config{}
	storage := NewStorage(gcfg)
	job := NewJob("test-job", *cfg, gcfg, storage)
	cmd, err := job.GetCmd()
	if err != nil {
		t.Fatal("Error:", err)
	}
	if strings.Join(cmd.Args, "|||") != "/usr/bin/ssh|||test-host.example|||-oBatchMode=yes|||-p|||22|||  TEST='oneone' TWO='xxx' /bin/bash" {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"))
	}
	t.Log(cmd.Args)
}

func TestJobGetCmdRemoteNoArgs(t *testing.T) {
	cfg := &JobConfig{
		Host: "test-host.example",
		Port: 2323,
	}
	gcfg := &Config{}
	storage := NewStorage(gcfg)
	job := NewJob("test-job", *cfg, gcfg, storage)
	cmd, err := job.GetCmd()
	if err != nil {
		t.Fatal("Error:", err)
	}
	if strings.Join(cmd.Args, "|||") != "/usr/bin/ssh|||test-host.example|||-oBatchMode=yes|||-p|||2323||| /bin/bash" {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"))
	}
	t.Log(cmd.Args)
}
