package bakapy

import (
	"fmt"
	"os"
	"os/user"
	"testing"
)

func TestSendFailedJobNotification_WithoutParameters(t *testing.T) {
	var cfg = &SMTPConfig{}

	meta := &Metadata{
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

	meta := &Metadata{
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

	meta := &Metadata{
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
