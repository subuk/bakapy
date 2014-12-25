package bakapy

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"testing"
)

func TestRunJob_MetadataCreated(t *testing.T) {
	gConfig := NewConfig()
	gConfig.Listen = "1.1.1.1:1234"

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

func TestRunJob_MetadataDirDoesNotExist(t *testing.T) {
	gConfig := NewConfig()
	gConfig.Listen = "1.1.1.1:1234"
	gConfig.MetadataDir = "/DOES_NOT_EXIST"
	gConfig.CommandDir, _ = ioutil.TempDir("", "")
	defer os.RemoveAll(gConfig.CommandDir)

	storage := NewStorage(gConfig)
	jConfig := &JobConfig{
		Command:  "wow.cmd",
		executor: &TestOkExecutor{},
	}
	os.Create(gConfig.CommandDir + "/" + "wow.cmd")
	metadataPath := RunJob("testjob", jConfig, gConfig, storage)
	_, err := LoadJobMetadata(metadataPath)
	if err == nil {
		t.Fatal("metadata loaded but not expected")
	}
}

func TestSendFailedJobNotification_WithoutParameters(t *testing.T) {
	var cfg = &SMTPConfig{}

	meta := &JobMetadata{
		JobName: "test J0b",
		Message: "testing message",
		Output:  []byte("some output"),
		Errput:  []byte("some errput"),
	}

	fn, rec := mockSend(nil)
	sender := &mailSender{cfg: cfg, send: fn}

	if err := sender.SendFailedJobNotification(meta); err != nil {
		t.Error("Notification sending failed:", err.Error())
	}

	curUser, err := user.Current()
	if err != nil {
		curUser = &user.User{"0", "0", "root", "root", "/root"}
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	sample := fmt.Sprintf(`From: bakapy@%s
To: %s
Subject: [bakapy] job 'test J0b' failed
Content-Type: text/plain;charset=utf8

Job  failed:
testing message

Output:
-----------------------------
some output
-----------------------------

Errput:
-----------------------------
some errput
-----------------------------
`, hostname, curUser.Name)
	if fmt.Sprintf("%s", rec.msg) != sample {
		t.Error("Email body does not match the sample")
	}
}

func TestSendFailedJobNotification_WithAllParameters(t *testing.T) {
	localUser := &user.User{"0", "0", "root", "root", "/root"}
	localEmail := fmt.Sprintf("%s@localhost", localUser.Username)
	fromEmail := "bakapy@localhost"
	var cfg = &SMTPConfig{"localhost", 25, fromEmail, localEmail}

	meta := &JobMetadata{
		JobName: "test J0b",
		Message: "testing message",
		Output:  []byte("some output"),
		Errput:  []byte("some errput"),
	}

	fn, rec := mockSend(nil)
	sender := &mailSender{cfg: cfg, send: fn}

	if err := sender.SendFailedJobNotification(meta); err != nil {
		t.Error("Notification sending failed:", err.Error())
	}

	sample := fmt.Sprintf(`From: bakapy@localhost
To: %s
Subject: [bakapy] job 'test J0b' failed
Content-Type: text/plain;charset=utf8

Job  failed:
testing message

Output:
-----------------------------
some output
-----------------------------

Errput:
-----------------------------
some errput
-----------------------------
`, localEmail)
	if fmt.Sprintf("%s", rec.msg) != sample {
		t.Error("Email body does not match the sample")
	}
}

func TestSendFailedJobNotification_WithOuterEmail(t *testing.T) {
	fromEmail := "bakapy@localhost"
	var cfg = &SMTPConfig{"localhost", 25, fromEmail, "bakapy@mailforspam.com"}

	meta := &JobMetadata{
		JobName: "test J0b",
		Message: "testing message",
		Output:  []byte("some output"),
		Errput:  []byte("some errput"),
	}

	fn, rec := mockSend(nil)
	sender := &mailSender{cfg: cfg, send: fn}

	if err := sender.SendFailedJobNotification(meta); err != nil {
		t.Error("Notification sending failed:", err.Error())
	}

	sample := `From: bakapy@localhost
To: bakapy@mailforspam.com
Subject: [bakapy] job 'test J0b' failed
Content-Type: text/plain;charset=utf8

Job  failed:
testing message

Output:
-----------------------------
some output
-----------------------------

Errput:
-----------------------------
some errput
-----------------------------
`
	if fmt.Sprintf("%s", rec.msg) != sample {
		t.Error("Email body does not match the sample")
	}
}

func mockSend(errToReturn error) (func(string, string, string, []byte) error, *emailRecorder) {
	rec := new(emailRecorder)
	return func(addr string, from string, to string, msg []byte) error {
		*rec = emailRecorder{addr, from, to, msg}
		return errToReturn
	}, rec
}

type emailRecorder struct {
	addr string
	from string
	to   string
	msg  []byte
}
