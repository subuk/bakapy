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
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(fmt.Errorf("cannot create temporary dir for test metaman:", err))
	}
	defer os.RemoveAll(tmpdir)

	n := NewScriptedNotificator(tmpdir, "does_not_exist", nil)
	md := Metadata{}
	err = n.JobFinished(md)
	if err == nil {
		t.Fatal("error expected")
	}
	expected := "fork/exec " + tmpdir + "/notify-does_not_exist.sh: no such file or directory"
	if err.Error() != expected {
		t.Fatal("bad error", err)
	}
}

func TestScriptedNotificatorJobFinished_ScriptOk(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(fmt.Errorf("cannot create temporary dir for test metaman:", err))
	}
	defer os.RemoveAll(tmpdir)
	ioutil.WriteFile(tmpdir+"/notify-test-exist.sh", TEST_SCRIPT, 0755)
	out := bytes.NewBuffer([]byte{})
	n := NewScriptedNotificator(tmpdir, "test-exist", map[string]string{
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
	err = n.JobFinished(md)
	if err != nil {
		t.Fatal("error", err)
	}
	expected := "\n"
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
