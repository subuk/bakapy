package bakapy

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestRunJob_MetadataCreated(t *testing.T) {
	gConfig := NewConfig()
	gConfig.Listen = "1.1.1.1:1234"
	gConfig.CommandDir = "/tmp/DoesNotExist"

	gConfig.MetadataDir, _ = ioutil.TempDir("", "")
	defer os.RemoveAll(gConfig.MetadataDir)

	gConfig.CommandDir, _ = ioutil.TempDir("", "")
	defer os.RemoveAll(gConfig.CommandDir)

	storage := NewStorage(gConfig)
	jConfig := &JobConfig{
		Command:  "wow.cmd",
		executor: &TestOkExecutor{},
	}
	os.Create(gConfig.CommandDir + "/" + "wow.cmd")
	metadataPath := RunJob("testjob", jConfig, gConfig, storage)
	meta, err := LoadJobMetadata(metadataPath)
	if err != nil {
		t.Fatal("cannot load metadata:", err)
	}
	if !meta.Success {
		t.Fatal()
	}
}
