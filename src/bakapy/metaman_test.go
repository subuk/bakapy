package bakapy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var ADDOK_EXPECTED_CONTENT = []byte(`{"JobName":"test","Gzip":false,"Namespace":"ns","TaskId":"123","Command":"cmd","Success":false,"Message":"","TotalSize":0,"StartTime":"2015-02-12T22:07:54.271257193Z","EndTime":"0001-01-01T00:00:00Z","ExpireTime":"2015-02-12T22:07:54.271258193Z","Files":null,"Pid":0,"RetCode":0,"Script":null,"Output":null,"Errput":null,"Config":{"Sudo":false,"Disabled":false,"Gzip":false,"MaxAgeDays":0,"MaxAge":0,"Namespace":"","Host":"","Port":0,"Command":"","Args":null,"RunAt":{"Second":"","Minute":"","Hour":"","Day":"","Month":"","Weekday":""}}}`)

func NewTestMetaMan() MetaManager {
	tmpdir, err := ioutil.TempDir("", "metamantest_")
	if err != nil {
		panic(fmt.Errorf("cannot create temporary dir for test metaman:", err))
	}
	return NewMetaMan(&Config{MetadataDir: tmpdir})
}

func TestMetaMan_AddOk(t *testing.T) {
	mm := NewTestMetaMan()
	defer os.RemoveAll(mm.(*MetaMan).RootDir)
	md := Metadata{
		JobName:   "test",
		Namespace: "ns",
		Command:   "cmd",
	}
	err := mm.Add("123", md)
	content, err := ioutil.ReadFile(mm.(*MetaMan).RootDir + "/123")
	if err != nil {
		t.Fatal("cannot add test entry:", err)
	}

	if bytes.Equal(content, ADDOK_EXPECTED_CONTENT) {
		t.Fatal("bad content", string(content))
	}

}

func TestMetaMan_AddAlreadyExist(t *testing.T) {
	mm := NewTestMetaMan()
	defer os.RemoveAll(mm.(*MetaMan).RootDir)
	md := Metadata{
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

func TestMetaMan_RemoveOk(t *testing.T) {
	mm := NewTestMetaMan()
	defer os.RemoveAll(mm.(*MetaMan).RootDir)
	f, err := os.Create(mm.(*MetaMan).RootDir + "/testfile")
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

func TestMetaMan_KeysNoDirectory(t *testing.T) {
	mm := NewTestMetaMan()
	os.RemoveAll(mm.(*MetaMan).RootDir)

	defer func() {
		expected := "cannot list metadata directory: open " + mm.(*MetaMan).RootDir + ": no such file or directory"
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

func TestMetaMan_KeysOk(t *testing.T) {
	mm := NewTestMetaMan()
	rootDir := mm.(*MetaMan).RootDir
	defer os.RemoveAll(rootDir)
	os.Create(rootDir + "/wow1")
	os.Create(rootDir + "/wow2")
	os.Create(rootDir + "/how")

	var result []TaskId
	for k := range mm.Keys() {
		result = append(result, k)
	}
	if len(result) != 3 {
		t.Fatal("bad result", result)
	}
}

func TestMetaMan_UpdateKeyDoesNotExist(t *testing.T) {
	mm := NewTestMetaMan()
	rootDir := mm.(*MetaMan).RootDir
	defer os.RemoveAll(rootDir)

	err := mm.Update("DOES_NOT_EXIST", func(m *Metadata) {
		t.Fatal("this block must not be called")
	})
	if err == nil {
		t.Fatal("error expected")
	}
	if err.Error() != "open "+rootDir+"/DOES_NOT_EXIST: no such file or directory" {
		t.Fatal("bad err", err)
	}

}

func TestMetaMan_UpdateKeyInvalidMetadata(t *testing.T) {
	mm := NewTestMetaMan()
	rootDir := mm.(*MetaMan).RootDir
	defer os.RemoveAll(rootDir)
	ioutil.WriteFile(rootDir+"/wow1", []byte("0asdf,'asd@34!"), 0666)

	err := mm.Update("wow1", func(m *Metadata) {
		t.Fatal("this block must not be called")
	})
	if err == nil {
		t.Fatal("error expected")
	}
	if err.Error() != "invalid character 'a' after top-level value" {
		t.Fatal("bad err", err)
	}

}
