package main

import (
	"bakapy"
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"errors"
	"github.com/op/go-logging"
	"io"
	"net"
	"testing"
)

type dummyAddr string

func (a dummyAddr) Network() string { return string(a) }
func (a dummyAddr) String() string  { return string(a) }

type DummyReader struct {
	data  []byte
	err   error
	shift int
}

func (r *DummyReader) Read(p []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	readEnd := r.shift + len(p)
	if readEnd > len(r.data) {
		readEnd = len(r.data)
	}

	if r.shift == readEnd {
		return 0, io.EOF
	}

	toCopy := r.data[r.shift:readEnd]
	copy(p, toCopy)
	r.shift += len(toCopy)

	return len(toCopy), nil
}

func (r *DummyReader) RemoteAddr() net.Addr {
	return dummyAddr("1.1.1.1")
}

func TestStorageConn_ReadTaskId_BadState(t *testing.T) {
	reader := &DummyReader{}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	conn.State = bakapy.STATE_END
	_, err := conn.ReadTaskId()
	if err == nil {
		t.Fatal("error not returned")
	}
	expectedError := "protocol error - cannot read task id in state 4"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorageConn_ReadTaskId_ErrorOnRead(t *testing.T) {
	reader := &DummyReader{
		err: errors.New("Oops"),
	}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	_, err := conn.ReadTaskId()
	if err == nil {
		t.Fatal("error not returned")
	}
	expectedError := "received error on reading task id: Oops"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorageConn_ReadTaskId_Ok(t *testing.T) {
	expectedTaskId := bakapy.TaskId(uuid.NewUUID().String())
	reader := &DummyReader{
		data: []byte(expectedTaskId),
	}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	taskId, err := conn.ReadTaskId()
	if err != nil {
		t.Fatal("error", err)
	}
	if taskId != expectedTaskId {
		t.Fatal("bad taskid:", taskId)
	}
	if conn.State != bakapy.STATE_WAIT_FILENAME {
		t.Fatal("conn.State must be ", bakapy.STATE_WAIT_FILENAME, "not", conn.State)
	}
}

func TestStorageConn_ReadFilename_BadState(t *testing.T) {
	reader := &DummyReader{}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	conn.State = bakapy.STATE_END
	_, err := conn.ReadFilename()
	if err == nil {
		t.Fatal("error not returned")
	}
	expectedError := "protocol error - cannot read filename in state 4"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorageConn_ReadFilename_ReadLenError(t *testing.T) {
	reader := &DummyReader{
		err: errors.New("Oops"),
	}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	conn.State = bakapy.STATE_WAIT_FILENAME
	_, err := conn.ReadFilename()
	if err == nil {
		t.Fatal("error not returned")
	}
	expectedError := "error while reading filename length: Oops"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorageConn_ReadFilename_ReadNameError(t *testing.T) {
	reader := &DummyReader{
		data: []byte("0117too_short"),
	}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	conn.State = bakapy.STATE_WAIT_FILENAME
	_, err := conn.ReadFilename()
	if err == nil {
		t.Fatal("error not returned")
	}
	expectedError := "cannot read filename: unexpected EOF"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorageConn_ReadFilename_ReadBadLen(t *testing.T) {
	reader := &DummyReader{
		data: []byte("hello"),
	}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	conn.State = bakapy.STATE_WAIT_FILENAME
	_, err := conn.ReadFilename()
	if err == nil {
		t.Fatal("error not returned")
	}
	expectedError := "cannot convert readed filename length to integer:hell: strconv.ParseInt: parsing \"hell\": invalid syntax"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorageConn_ReadFilename_Ok(t *testing.T) {
	expectedFilename := "mypath/tofile.txt"
	reader := &DummyReader{
		data: []byte("0017" + expectedFilename),
	}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	conn.State = bakapy.STATE_WAIT_FILENAME
	filename, err := conn.ReadFilename()

	if err != nil {
		t.Fatal("error", err)
	}

	if filename != expectedFilename {
		t.Fatalf("Bad filename '%s' expected '%s'", []byte(filename), []byte(expectedFilename))
	}
	if conn.State != bakapy.STATE_WAIT_DATA {
		t.Fatal("conn.State must be ", bakapy.STATE_WAIT_DATA, "not", conn.State)
	}
}

func TestStorageConn_ReadContent_BadState(t *testing.T) {
	reader := &DummyReader{}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	conn.State = bakapy.STATE_END

	output := bytes.NewBuffer([]byte(""))

	_, err := conn.ReadContent(output)
	if err == nil {
		t.Fatal("error not returned")
	}
	expectedError := "protocol error - cannot read data in state 4"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorageConn_ReadContent_ReadError(t *testing.T) {
	reader := &DummyReader{
		err: errors.New("oops"),
	}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	conn.State = bakapy.STATE_WAIT_DATA

	output := bytes.NewBuffer([]byte(""))

	_, err := conn.ReadContent(output)
	if err == nil {
		t.Fatal("error not returned")
	}
	expectedError := "read file content error: oops"
	if err.Error() != expectedError {
		t.Fatal("bad error:", err)
	}
}

func TestStorageConn_ReadContent_Ok(t *testing.T) {
	reader := &DummyReader{
		data: []byte("such content"),
	}
	conn := NewStorageConn(reader, logging.MustGetLogger("connection.test"))
	conn.State = bakapy.STATE_WAIT_DATA

	output := bytes.NewBuffer(make([]byte, 1000))

	written, err := conn.ReadContent(output)
	if err != nil {
		t.Fatal("error", err)
	}

	if int(written) != len(reader.data) {
		t.Fatal("written != len(reader.data)", written, "!=", len(reader.data))
	}

	if conn.State != bakapy.STATE_END {
		t.Fatal("conn.State must be ", bakapy.STATE_END, "not", conn.State)
	}
}
