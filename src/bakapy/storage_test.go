package bakapy

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"testing"
)

type NullStorageProtocol struct {
	readContentCalled bool
	filename          string
	content           []byte
	taskId            string
}

func (p *NullStorageProtocol) ReadTaskId() (TaskId, error) {
	return TaskId(p.taskId), nil
}
func (p *NullStorageProtocol) ReadFilename() (string, error) { return p.filename, nil }
func (p *NullStorageProtocol) ReadContent(output io.Writer) (int64, error) {
	p.readContentCalled = true
	output.Write(p.content)
	return int64(len(p.content)), nil
}
func (p *NullStorageProtocol) RemoteAddr() net.Addr { return dummyAddr("1.1.1.1") }

func TestStorage_HandleConnection_UnknownTaskId(t *testing.T) {
	protohandle := &NullStorageProtocol{taskId: "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c"}
	cfg := NewConfig()
	storage := NewStorage(cfg, NewTestMetaMan())
	err := storage.HandleConnection(protohandle)
	expectedError := "Cannot find task id 'a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c' in current job list, closing connection"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorage_HandleConnection_JobFinishWordWorks(t *testing.T) {
	protohandle := &NullStorageProtocol{
		taskId:   "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c",
		filename: JOB_FINISH,
	}
	cfg := NewConfig()
	cfg.StorageDir, _ = ioutil.TempDir("", "test_bakapy_storage")
	defer os.RemoveAll(cfg.StorageDir)
	storage := NewStorage(cfg, NewTestMetaMan())
	err := storage.metaman.Add("testjob", "test/wow", "cmd", "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", true, 10000)
	if err != nil {
		t.Fatal("cannot add metadata:", err)
	}
	err = storage.HandleConnection(protohandle)
	if err != nil {
		t.Fatal("unexpected error ", err)
	}
	if protohandle.readContentCalled {
		t.Fatal("file content was readed")
	}
}

func TestStorage_HandleConnection_SaveGzip(t *testing.T) {
	protohandle := &NullStorageProtocol{
		taskId:   "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c",
		filename: "hello.txt",
		content:  []byte("testcontent"),
	}
	cfg := NewConfig()
	cfg.StorageDir, _ = ioutil.TempDir("", "test_bakapy_storage")
	defer os.RemoveAll(cfg.StorageDir)
	storage := NewStorage(cfg, NewTestMetaMan())

	err := storage.metaman.Add("testjob", "test/wow", "xxx", "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", true, 10000)
	if err != nil {
		t.Fatal("cannot add metadata:", err)
	}

	err = storage.HandleConnection(protohandle)
	if err != nil {
		t.Fatal("error", err)
	}

	expectedFilePath := path.Join(cfg.StorageDir, "test/wow", protohandle.filename+".gz")
	file, err := os.Open(expectedFilePath)
	if err != nil {
		t.Fatal("expected file open error:", err)
	}
	gzFile, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	fileContent, err := ioutil.ReadAll(gzFile)
	if err != nil {
		t.Fatal("read file content error:", err)
	}

	if string(fileContent) != "testcontent" {
		t.Fatal("unexpected file content", fileContent)
	}
}

func TestStorage_HandleConnection_SaveNotGzip(t *testing.T) {
	protohandle := &NullStorageProtocol{
		taskId:   "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c",
		filename: "world.txt",
		content:  []byte("test_ungz_content"),
	}
	cfg := NewConfig()
	cfg.StorageDir, _ = ioutil.TempDir("", "test_bakapy_storage")
	defer os.RemoveAll(cfg.StorageDir)
	storage := NewStorage(cfg, NewTestMetaMan())

	err := storage.metaman.Add("testjob", "test/wow", "xxx", "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", false, 10000)
	if err != nil {
		t.Fatal("cannot add metadata:", err)
	}

	err = storage.HandleConnection(protohandle)
	if err != nil {
		t.Fatal("error", err)
	}

	expectedFilePath := path.Join(cfg.StorageDir, "test/wow", protohandle.filename)
	fileContent, err := ioutil.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatal("expected file read error:", err)
	}

	if string(fileContent) != "test_ungz_content" {
		t.Fatal("unexpected file content", string(fileContent))
	}
}

// func TestStorage_HandleConnection_MetadataSended(t *testing.T) {
// 	protohandle := &NullStorageProtocol{
// 		filename: "hello.txt",
// 		content:  []byte("wow"),
// 	}
// 	cfg := NewConfig()
// 	cfg.StorageDir, _ = ioutil.TempDir("", "test_bakapy_storage")
// 	defer os.RemoveAll(cfg.StorageDir)
// 	storage := NewStorage(cfg, NewTestMetaMan())

// 	err := storage.HandleConnection(protohandle)
// 	if err != nil {
// 		t.Fatal("error", err)
// 	}

// 	if len(cJob.FileAddChan) != 1 {
// 		t.Fatal("number of files in fileAddChan is ", len(cJob.FileAddChan))
// 	}

// 	fileMeta := <-cJob.FileAddChan

// 	if fileMeta.Name != "hello.txt" {
// 		t.Fatal("bad filename", fileMeta.Name)
// 	}
// 	if fileMeta.Size != int64(len(protohandle.content)) {
// 		t.Fatal("bad file size", fileMeta.Size)
// 	}
// 	if fileMeta.SourceAddr != "1.1.1.1" {
// 		t.Fatal("bad source address", fileMeta.SourceAddr)
// 	}
// 	if fileMeta.StartTime == (time.Time{}) {
// 		t.Fatal("bad start time", fileMeta.StartTime)
// 	}
// 	if fileMeta.EndTime == (time.Time{}) {
// 		t.Fatal("bad end time", fileMeta.EndTime)
// 	}
// }
