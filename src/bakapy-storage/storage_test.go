package main

import (
	"bakapy"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"testing"
	"time"
)

type TestMetaMan struct {
	stor   map[bakapy.TaskId]bakapy.Metadata
	addErr error
}

func NewTestMockMetaMan() *TestMetaMan {
	return &TestMetaMan{
		stor: make(map[bakapy.TaskId]bakapy.Metadata),
	}
}

func (mm *TestMetaMan) Keys() chan bakapy.TaskId {
	ch := make(chan bakapy.TaskId)
	go func() {
		for key, _ := range mm.stor {
			ch <- key
		}
		close(ch)
	}()
	return ch
}

func (mm *TestMetaMan) View(id bakapy.TaskId) (bakapy.Metadata, error) {
	md, ok := mm.stor[id]
	if !ok {
		return bakapy.Metadata{}, errors.New("does not exist")
	}
	return md, nil
}

func (mm *TestMetaMan) Add(id bakapy.TaskId, md bakapy.Metadata) error {
	md.TaskId = id
	if mm.addErr == nil {
		mm.stor[id] = md
		return nil
	}
	return mm.addErr
}

func (mm *TestMetaMan) Update(id bakapy.TaskId, up func(*bakapy.Metadata)) error {
	md, err := mm.View(id)
	if err != nil {
		return err
	}
	up(&md)
	mm.stor[id] = md
	return nil
}

func (mm *TestMetaMan) Remove(id bakapy.TaskId) error {
	delete(mm.stor, id)
	return nil
}

func (mm *TestMetaMan) AddFile(id bakapy.TaskId, fm bakapy.MetadataFileEntry) error {
	md := mm.stor[id]
	md.Files = append(md.Files, fm)
	mm.stor[id] = md
	return nil
}

type NullStorageProtocol struct {
	readContentCalled bool
	filename          string
	content           []byte
	taskId            string
	readContentErr    error
}

func (p *NullStorageProtocol) ReadTaskId() (bakapy.TaskId, error) {
	return bakapy.TaskId(p.taskId), nil
}
func (p *NullStorageProtocol) ReadFilename() (string, error) { return p.filename, nil }
func (p *NullStorageProtocol) ReadContent(output io.Writer) (int64, error) {
	p.readContentCalled = true
	output.Write(p.content)
	return int64(len(p.content)), p.readContentErr
}
func (p *NullStorageProtocol) RemoteAddr() net.Addr { return dummyAddr("1.1.1.1") }

type NullStorageProtocolErrorReadTaskId struct {
	NullStorageProtocol
}

func (p *NullStorageProtocolErrorReadTaskId) ReadTaskId() (bakapy.TaskId, error) {
	return bakapy.TaskId(""), errors.New("test error")
}

type NullStorageProtocolErrorReadFilename struct {
	NullStorageProtocol
}

func (p *NullStorageProtocolErrorReadFilename) ReadFilename() (string, error) {
	return "", errors.New("filename test error")
}

func TestStorage_HandleConnection_TaskIdReadErr(t *testing.T) {
	protohandle := &NullStorageProtocolErrorReadTaskId{}
	storage := NewStorage("", "", NewTestMockMetaMan())
	err := storage.HandleConnection(protohandle)
	if err == nil {
		t.Fatal("error expected")
	}
	expectedError := "cannot read task id: test error. closing connection"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorage_HandleConnection_UnknownTaskId(t *testing.T) {
	protohandle := &NullStorageProtocol{taskId: "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c"}
	storage := NewStorage("", "", NewTestMockMetaMan())
	err := storage.HandleConnection(protohandle)
	if err == nil {
		t.Fatal("error expected")
	}
	expectedError := "cannot find task id a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c: does not exist"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorage_HandleConnection_TaskAlreadyFinished(t *testing.T) {
	protohandle := &NullStorageProtocol{taskId: "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c"}
	storage := NewStorage("", "", NewTestMockMetaMan())

	md := bakapy.Metadata{
		JobName:   "testjob",
		Namespace: "test/wow",
		Command:   "cmd",
		EndTime:   time.Now(),
	}
	err := storage.metaman.Add("a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", md)
	if err != nil {
		t.Fatal("cannot add metadata:", err)
	}

	err = storage.HandleConnection(protohandle)
	if err == nil {
		t.Fatal("error expected")
	}
	expectedError := "task with id 'a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c' already finished, closing connection"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorage_HandleConnection_FilenameReadErr(t *testing.T) {
	protohandle := &NullStorageProtocolErrorReadFilename{NullStorageProtocol{taskId: "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c"}}
	storage := NewStorage("", "", NewTestMockMetaMan())

	md := bakapy.Metadata{
		JobName:   "testjob",
		Namespace: "test/wow",
		Command:   "cmd",
	}
	err := storage.metaman.Add("a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", md)
	if err != nil {
		t.Fatal("cannot add metadata:", err)
	}

	err = storage.HandleConnection(protohandle)
	if err == nil {
		t.Fatal("error expected")
	}
	expectedError := "cannot read filename: filename test error. closing connection"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorage_HandleConnection_JobFinishWordWorks(t *testing.T) {
	protohandle := &NullStorageProtocol{
		taskId:   "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c",
		filename: bakapy.JOB_FINISH,
	}
	storageDir, _ := ioutil.TempDir("", "test_bakapy_storage")
	defer os.RemoveAll(storageDir)
	storage := NewStorage(storageDir, "", NewTestMockMetaMan())
	md := bakapy.Metadata{
		JobName:   "testjob",
		Namespace: "test/wow",
		Command:   "cmd",
	}
	err := storage.metaman.Add("a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", md)
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
	storageDir, _ := ioutil.TempDir("", "test_bakapy_storage")
	defer os.RemoveAll(storageDir)
	storage := NewStorage(storageDir, "", NewTestMockMetaMan())
	md := bakapy.Metadata{
		JobName:   "testjob",
		Namespace: "test/wow",
		Command:   "xxx",
		Gzip:      true,
	}
	err := storage.metaman.Add("a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", md)
	if err != nil {
		t.Fatal("cannot add metadata:", err)
	}

	err = storage.HandleConnection(protohandle)
	if err != nil {
		t.Fatal("error", err)
	}

	expectedFilePath := path.Join(storageDir, "test/wow", protohandle.filename+".gz")
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
	storageDir, _ := ioutil.TempDir("", "test_bakapy_storage")
	defer os.RemoveAll(storageDir)
	storage := NewStorage(storageDir, "", NewTestMockMetaMan())
	md := bakapy.Metadata{
		JobName:   "testjob",
		Namespace: "test/wow",
		Command:   "xxx",
		Gzip:      false,
	}
	err := storage.metaman.Add("a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", md)
	if err != nil {
		t.Fatal("cannot add metadata:", err)
	}

	err = storage.HandleConnection(protohandle)
	if err != nil {
		t.Fatal("error", err)
	}

	expectedFilePath := path.Join(storageDir, "test/wow", protohandle.filename)
	fileContent, err := ioutil.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatal("expected file read error:", err)
	}

	if string(fileContent) != "test_ungz_content" {
		t.Fatal("unexpected file content", string(fileContent))
	}
}

func TestStorage_HandleConnection_DestDirsMakeFailed(t *testing.T) {
	protohandle := &NullStorageProtocol{
		taskId:   "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c",
		filename: "world.txt",
	}
	storageDir, _ := ioutil.TempDir("", "test_bakapy_storage")
	defer os.RemoveAll(storageDir)
	storage := NewStorage(storageDir, "", NewTestMockMetaMan())
	md := bakapy.Metadata{
		JobName:   "testjob",
		Namespace: "test/wow",
		Command:   "xxx",
		Gzip:      false,
	}
	err := storage.metaman.Add("a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", md)
	if err != nil {
		t.Fatal("cannot add metadata:", err)
	}

	f, err := os.Create(storage.RootDir + "/test")
	if err != nil {
		t.Fatal("cannot create test file:", err)
	}
	f.Close()

	err = storage.HandleConnection(protohandle)
	if err == nil {
		t.Fatal("error expected")
	}
	expected := "cannot create file folder: mkdir " + storage.RootDir + "/test: not a directory"
	if err.Error() != expected {
		t.Fatal("bad err", err)
	}
}

func TestStorage_HandleConnection_DestFileOpenFailed(t *testing.T) {
	protohandle := &NullStorageProtocol{
		taskId:   "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c",
		filename: "world.txt",
	}
	storageDir, _ := ioutil.TempDir("", "test_bakapy_storage")
	defer os.RemoveAll(storageDir)
	storage := NewStorage(storageDir, "", NewTestMockMetaMan())
	md := bakapy.Metadata{
		JobName:   "testjob",
		Namespace: "test/wow",
		Command:   "xxx",
		Gzip:      false,
	}
	err := storage.metaman.Add("a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", md)
	if err != nil {
		t.Fatal("cannot add metadata:", err)
	}

	err = os.MkdirAll(storage.RootDir+"/test/wow/world.txt", 0755)
	if err != nil {
		t.Fatal("cannot create test file:", err)
	}

	err = storage.HandleConnection(protohandle)
	if err == nil {
		t.Fatal("error expected")
	}
	expected := "cannot open file: open " + storage.RootDir + "/test/wow/world.txt: is a directory"
	if err.Error() != expected {
		t.Fatal("bad err", err)
	}
}

func TestStorage_HandleConnection_ReadContentFailed(t *testing.T) {
	protohandle := &NullStorageProtocol{
		taskId:         "a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c",
		filename:       "world.txt",
		readContentErr: errors.New("test err"),
	}
	storageDir, _ := ioutil.TempDir("", "test_bakapy_storage")
	defer os.RemoveAll(storageDir)
	storage := NewStorage(storageDir, "", NewTestMockMetaMan())
	md := bakapy.Metadata{
		JobName:   "testjob",
		Namespace: "test/wow",
		Command:   "xxx",
		Gzip:      false,
	}
	err := storage.metaman.Add("a70cb394-c22d-4fe7-a5cc-bc0a5e19a24c", md)
	if err != nil {
		t.Fatal("cannot add metadata:", err)
	}

	err = storage.HandleConnection(protohandle)
	if err == nil {
		t.Fatal("error expected")
	}
	expected := "cannot save file: test err. closing connection"
	if err.Error() != expected {
		t.Fatal("bad err", err)
	}
}
