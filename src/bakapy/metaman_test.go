package bakapy

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

var ADDOK_EXPECTED_CONTENT = []byte(`{"JobName":"test","Gzip":false,"Namespace":"ns","TaskId":"123","Command":"cmd","Success":false,"Message":"","TotalSize":0,"StartTime":"2015-02-12T22:07:54.271257193Z","EndTime":"0001-01-01T00:00:00Z","ExpireTime":"2015-02-12T22:07:54.271258193Z","Files":null,"Pid":0,"RetCode":0,"Script":null,"Output":null,"Errput":null,"Config":{"Sudo":false,"Disabled":false,"Gzip":false,"MaxAgeDays":0,"MaxAge":0,"Namespace":"","Host":"","Port":0,"Command":"","Args":null,"RunAt":{"Second":"","Minute":"","Hour":"","Day":"","Month":"","Weekday":""}}}`)

func TestMetaMan_AddOk(t *testing.T) {
	mm := NewTestMetaMan()
	defer os.RemoveAll(mm.(*MetaMan).RootDir)
	err := mm.Add("test", "ns", "cmd", "123", false, 1000)
	content, err := ioutil.ReadFile(mm.(*MetaMan).RootDir + "/123")
	if err != nil {
		t.Fatal("cannot add test entry:", err)
	}

	if bytes.Equal(content, ADDOK_EXPECTED_CONTENT) {
		t.Fatal("bad content", string(content))
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
