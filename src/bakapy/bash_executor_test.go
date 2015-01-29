package bakapy

import (
	"bytes"
	"strings"
	"testing"
)

func TestBashExecutor_GetCmd_Local(t *testing.T) {
	args := map[string]string{
		"test": "oneone",
	}
	host := ""
	port := uint(22)
	sudo := false
	executor := NewBashExecutor(args, host, port, sudo, "", nil)
	cmd, err := executor.GetCmd()

	if err != nil {
		t.Fatal("Error:", err)
	}
	if strings.Join(cmd.Args, "|||") != "/bin/bash|||-c||| TEST='oneone' /bin/bash" {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"), "expected '/bin/bash|||-c||| TEST='oneone' /bin/bash'")
	}
	t.Log(cmd.Args)
}

func TestBashExecutor_GetCmd_LocalSudo(t *testing.T) {
	args := map[string]string{
		"test": "oneone",
	}
	host := ""
	port := uint(22)
	sudo := true
	executor := NewBashExecutor(args, host, port, sudo, "", nil)
	cmd, err := executor.GetCmd()
	if err != nil {
		t.Fatal("Error:", err)
	}
	expected_args := "/bin/bash|||-c|||sudo  TEST='oneone' /bin/bash"
	if strings.Join(cmd.Args, "|||") != expected_args {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"), "expected", expected_args)
	}
	t.Log(cmd.Args)
}

func TestBashExecutor_GetCmd_Remote(t *testing.T) {
	args := map[string]string{
		"test": "oneone",
	}
	host := "test-host.example"
	port := uint(2424)
	sudo := false
	executor := NewBashExecutor(args, host, port, sudo, "", nil)
	cmd, err := executor.GetCmd()

	if err != nil {
		t.Fatal("Error:", err)
	}
	expected_args := "/usr/bin/ssh|||test-host.example|||-oBatchMode=yes|||-p|||2424||| TEST='oneone' /bin/bash"
	if strings.Join(cmd.Args, "|||") != expected_args {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"), "expected", expected_args)
	}
	t.Log(cmd.Args)
}

func TestBashExecutor_GetCmd_RemoteSudo(t *testing.T) {
	args := map[string]string{
		"test": "oneone",
	}
	host := "test-host.example"
	port := uint(2424)
	sudo := true
	executor := NewBashExecutor(args, host, port, sudo, "", nil)
	cmd, err := executor.GetCmd()

	if err != nil {
		t.Fatal("Error:", err)
	}
	expected_args := "/usr/bin/ssh|||test-host.example|||-oBatchMode=yes|||-p|||2424|||sudo  TEST='oneone' /bin/bash"
	if strings.Join(cmd.Args, "|||") != expected_args {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"), "expected", expected_args)
	}
	t.Log(cmd.Args)
}

func TestBashExecutor_GetCmd_RemoteNoPort(t *testing.T) {
	args := map[string]string{
		"test": "oneone",
	}
	host := "test-host.example"
	port := uint(0)
	sudo := false
	executor := NewBashExecutor(args, host, port, sudo, "", nil)
	cmd, err := executor.GetCmd()

	if err != nil {
		t.Fatal("Error:", err)
	}
	expected_args := "/usr/bin/ssh|||test-host.example|||-oBatchMode=yes|||-p|||22||| TEST='oneone' /bin/bash"
	if strings.Join(cmd.Args, "|||") != expected_args {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"), "expected", expected_args)
	}
	t.Log(cmd.Args)
}

func TestBashExecutor_GetCmd_RemoteNoArgs(t *testing.T) {
	args := map[string]string{}
	host := "test-host.example"
	port := uint(2323)
	sudo := false
	executor := NewBashExecutor(args, host, port, sudo, "", nil)
	cmd, err := executor.GetCmd()

	if err != nil {
		t.Fatal("Error:", err)
	}
	if strings.Join(cmd.Args, "|||") != "/usr/bin/ssh|||test-host.example|||-oBatchMode=yes|||-p|||2323||| /bin/bash" {
		t.Fatal("Wrong cmd args:", strings.Join(cmd.Args, "|||"))
	}
	t.Log(cmd.Args)
}

func TestBashExecutor_Execute_CommandOk(t *testing.T) {
	args := map[string]string{}
	host := ""
	port := uint(2323)
	sudo := false
	executor := NewBashExecutor(args, host, port, sudo, "", nil)

	script := []byte(`echo -n hello; exit 0;`)
	output := new(bytes.Buffer)
	errput := new(bytes.Buffer)
	err := executor.Execute(script, output, errput)
	if err != nil {
		t.Fatal("Error:", err)
	}
	if output.String() != "hello" {
		t.Fatalf("Output must be 'hello', not '%s'", output)
	}
	if errput.String() != "" {
		t.Fatalf("Errput must be '', not '%s'", errput)
	}
}

func TestBashExecutor_Execute_CommandFailed(t *testing.T) {
	args := map[string]string{}
	host := ""
	port := uint(2323)
	sudo := false
	executor := NewBashExecutor(args, host, port, sudo, "", nil)

	script := []byte(`echo -n some errput >&2; exit 19;`)
	output := new(bytes.Buffer)
	errput := new(bytes.Buffer)
	err := executor.Execute(script, output, errput)
	if err.Error() != "exit status 19" {
		t.Fatalf("err must be 'exit status 19', not '%s'", err)
	}

	if errput.String() != "some errput" {
		t.Fatalf("Errput must be 'some errput', not '%s'", errput)
	}
}
