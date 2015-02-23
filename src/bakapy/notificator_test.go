package bakapy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var TEST_SCRIPT = []byte(`#!/bin/bash
echo
for name in $(env |grep ^BAKAPY_ |awk -F= '{print $1}'|sort);do
    echo ${name}=$(eval echo '$'"$(echo $name)")
done
`)

func TestScriptedNotificatorJobFinished_ScriptNotFound(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	tmpfile.Close()
	spool := &TestScriptPool{fmt.Errorf("no script"), nil, tmpfile.Name()}
	n := NewScriptedNotificator(spool, "does_not_exist", nil)
	md := Metadata{}
	err := n.JobFinished(md)
	if err == nil {
		t.Fatal("error expected")
	}
	expected := "cannot get script does_not_exist: no script"
	if err.Error() != expected {
		t.Fatal("bad error", err)
	}
}

func TestScriptedNotificatorJobFinished_ScriptOk(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	tmpfile.Close()
	spool := &TestScriptPool{nil, TEST_SCRIPT, tmpfile.Name()}
	out := bytes.NewBuffer([]byte{})
	n := NewScriptedNotificator(spool, "test-exist", map[string]string{
		"param1": "wow",
		"pArom":  "hello",
	})
	n.output = out
	n.errput = out
	md := Metadata{
		Errput:  []byte(`some errput`),
		Output:  []byte(`some output`),
		JobName: "testjob",
		Message: "test message",
		TaskId:  TaskId("123123"),
	}
	err := n.JobFinished(md)
	if err != nil {
		t.Fatal("error", err)
	}
	expected := "\n"
	expected += "BAKAPY_EVENT=job_finished\n"
	expected += "BAKAPY_METADATA_ERRPUT=some errput\n"
	expected += "BAKAPY_METADATA_JOBNAME=testjob\n"
	expected += "BAKAPY_METADATA_MESSAGE=test message\n"
	expected += "BAKAPY_METADATA_OUTPUT=some output\n"
	expected += "BAKAPY_METADATA_SUCCESS=0\n"
	expected += "BAKAPY_METADATA_TASKID=123123\n"
	expected += "BAKAPY_PARAM_PARAM1=wow\n"
	expected += "BAKAPY_PARAM_PAROM=hello\n"

	if o := out.String(); o != expected {
		t.Fatal("bad script output", o)
	}
}

func TestScriptedNotificatorJobFinished_ScriptDeleted(t *testing.T) {
	tmpfile, _ := ioutil.TempFile("", "")
	tmpfile.Close()
	spool := &TestScriptPool{nil, TEST_SCRIPT, tmpfile.Name()}
	n := NewScriptedNotificator(spool, "test-exist", nil)
	md := Metadata{}
	err := n.JobFinished(md)
	if err != nil {
		t.Fatal("error", err)
	}
	if fi, err := os.Stat(tmpfile.Name()); err == nil {
		t.Fatal("temporary script file still exist:", fi.Name())
	}
}
