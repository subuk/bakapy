package main

import (
	"bakapy"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var ADDOK_EXPECTED_CONTENT = []byte(`{"JobName":"test","Gzip":false,"Namespace":"ns","TaskId":"123","Command":"cmd","Success":false,"Message":"","TotalSize":0,"StartTime":"2015-02-12T22:07:54.271257193Z","EndTime":"0001-01-01T00:00:00Z","ExpireTime":"2015-02-12T22:07:54.271258193Z","Files":null,"Pid":0,"RetCode":0,"Script":null,"Output":null,"Errput":null,"Config":{"Sudo":false,"Disabled":false,"Gzip":false,"MaxAgeDays":0,"MaxAge":0,"Namespace":"","Host":"","Port":0,"Command":"","Args":null,"RunAt":{"Second":"","Minute":"","Hour":"","Day":"","Month":"","Weekday":""}}}`)

func NewTestJSONDir() *JSONDir {
	tmpdir, err := ioutil.TempDir("", "metamantest_")
	if err != nil {
		panic(fmt.Errorf("cannot create temporary dir for test metaman:", err))
	}
	return NewJSONDir(tmpdir)
}

func TestJSONDir_AddOk(t *testing.T) {
	mm := NewTestJSONDir()
	defer os.RemoveAll(mm.RootDir)
	md := bakapy.Metadata{
		JobName:   "test",
		Namespace: "ns",
		Command:   "cmd",
	}
	err := mm.Add("123", md)
	content, err := ioutil.ReadFile(mm.RootDir + "/123")
	if err != nil {
		t.Fatal("cannot add test entry:", err)
	}

	if bytes.Equal(content, ADDOK_EXPECTED_CONTENT) {
		t.Fatal("bad content", string(content))
	}

}

func TestJSONDir_AddAlreadyExist(t *testing.T) {
	mm := NewTestJSONDir()
	defer os.RemoveAll(mm.RootDir)
	md := bakapy.Metadata{
		JobName:   "test",
		Namespace: "ns",
		Command:   "cmd",
	}
	mm.Add("123", md)
	err := mm.Add("123", md)
	if err == nil {
		t.Fatal("error expected")
	}
	if err.Error() != "metadata for task 123 already exist" {
		t.Fatal("bad error", err)
	}

}

func TestJSONDir_RemoveOk(t *testing.T) {
	mm := NewTestJSONDir()
	defer os.RemoveAll(mm.RootDir)
	f, err := os.Create(mm.RootDir + "/testfile")
	if err != nil {
		t.Fatal("cannot create test file:", err)
	}
	f.Write([]byte("{}"))
	f.Close()

	err = mm.Remove("testfile")
	if err != nil {
		t.Fatal("remove failed:", err)
	}
}

func TestJSONDir_KeysNoDirectory(t *testing.T) {
	mm := NewTestJSONDir()
	os.RemoveAll(mm.RootDir)

	defer func() {
		expected := "cannot list metadata directory: open " + mm.RootDir + ": no such file or directory"
		if r := recover(); r != nil {
			if err := r.(error); err.Error() != expected {
				t.Fatal("bad err", err)
			}
		} else {
			t.Fatal("panic expected")
		}
	}()
	mm.Keys()
}

func TestJSONDir_KeysOk(t *testing.T) {
	mm := NewTestJSONDir()
	rootDir := mm.RootDir
	defer os.RemoveAll(rootDir)
	os.Create(rootDir + "/wow1")
	os.Create(rootDir + "/wow2")
	os.Create(rootDir + "/how")

	var result []bakapy.TaskId
	for k := range mm.Keys() {
		result = append(result, k)
	}
	if len(result) != 3 {
		t.Fatal("bad result", result)
	}
}
