package bakapy

import (
	"io/ioutil"
	"os"
	"regexp"
	"testing"
)

func TestDirectoryScriptPool_BackupScript_Ok(t *testing.T) {
	root, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(root)

	spool := NewDirectoryScriptPool(&Config{CommandDir: root})
	f, _ := os.Create(root + "/wowfile")
	f.Write([]byte("Hello"))
	f.Close()
	script, err := spool.BackupScript("wowfile")
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if string(script) != "Hello" {
		t.Fatal("bad script content", script)
	}
}

func TestDirectoryScriptPool_BackupScript_NoFile(t *testing.T) {
	spool := NewDirectoryScriptPool(&Config{CommandDir: "/does_not_exist"})
	_, err := spool.BackupScript("wowfile")
	if err == nil {
		t.Fatal("error expected")
	}
	if err.Error() != "open /does_not_exist/wowfile: no such file or directory" {
		t.Fatal("bad error", err)
	}

}

func TestDirectoryScriptPool_NotifyScriptPath_ModeOk(t *testing.T) {
	root, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(root)
	spool := NewDirectoryScriptPool(&Config{CommandDir: root})
	f, _ := os.Create(root + "/notify-wowfile.sh")
	f.Write([]byte("Hello"))
	f.Close()
	scriptPath, err := spool.NotifyScriptPath("wowfile")
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	fi, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatal("cannot stat returned script path: %s", err)
	}
	mode := fi.Mode()
	if ms := mode.String(); ms != "-rwx------" {
		t.Fatal("bad mode", ms)
	}
}

func TestDirectoryScriptPool_NotifyScriptPath_NoScript(t *testing.T) {
	spool := NewDirectoryScriptPool(&Config{CommandDir: "/does_not_exist"})
	_, err := spool.NotifyScriptPath("wowfile")
	if err == nil {
		t.Fatal("error expected ")
	}
	if err.Error() != "open /does_not_exist/notify-wowfile.sh: no such file or directory" {
		t.Fatal("bad error", err)
	}
}

func TestDirectoryScriptPool_NotifyScriptPath_BadTempdir(t *testing.T) {
	root, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(root)
	spool := NewDirectoryScriptPool(&Config{CommandDir: root})
	spool.tmp = "/tmp_does_not_exist"
	f, _ := os.Create(root + "/notify-wowfile.sh")
	f.Write([]byte("Hello"))
	f.Close()
	_, err := spool.NotifyScriptPath("wowfile")
	if err == nil {
		t.Fatal("error expected")
	}
	expected := regexp.MustCompile(`^cannot create temp file for notify script wowfile: open /tmp_does_not_exist/notify-scriptwowfile[^:]+: no such file or directory$`)
	if !expected.MatchString(err.Error()) {
		t.Fatal("bad error", err)
	}
}
