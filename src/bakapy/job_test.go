package bakapy

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type TestOkExecutor struct{}

func (e *TestOkExecutor) Execute(script []byte, output io.Writer, errput io.Writer) error {
	return nil
}

type TestFailExecutor struct{}

func (e *TestFailExecutor) Execute(script []byte, output io.Writer, errput io.Writer) error {
	return errors.New("Oops")
}

func TestJob_Run_ExecutionOkMetadataSetted(t *testing.T) {
	now := time.Now()
	executor := &TestOkExecutor{}
	maxAge, _ := time.ParseDuration("30m")
	cfg := &JobConfig{Command: "utils.go", Namespace: "wow", Gzip: true, MaxAge: maxAge}
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("cannot create temp dir:", err)
	}
	gcfg := &Config{MetadataDir: tmpdir}
	metaman := NewMetaMan(gcfg)
	defer os.RemoveAll(gcfg.MetadataDir)
	job := NewJob(
		"test", cfg, "127.0.0.1:9999",
		".", executor, metaman,
	)

	err = job.Run()
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	m, err := metaman.View(job.TaskId)
	if err != nil {
		t.Fatal("cannot get job metadata:", err)
	}

	if !m.Success {
		t.Fatal("m.Success must be true")
	}

	if !m.Gzip {
		t.Fatal("m.Gzip must be true")
	}

	if m.Message != "OK" {
		t.Fatal("m.JobName must be 'OK' not", m.Message)
	}

	if m.JobName != "test" {
		t.Fatal("m.JobName must be 'test' not", m.JobName)
	}

	if m.Namespace != "wow" {
		t.Fatal("m.Namespace must be 'wow' not", m.Namespace)
	}

	if m.TaskId != job.TaskId {
		t.Fatalf("m.TaskId must be '%s' not '%s'", m.TaskId, job.TaskId)
	}

	if m.StartTime.Before(now) {
		t.Fatalf("m.StartTime before ", now)
	}

	if m.EndTime.Before(now) {
		t.Fatalf("m.EndTime before ", now)
	}

	expected_expire := m.StartTime.Add(maxAge)
	if !m.ExpireTime.Equal(expected_expire) {
		t.Fatalf("m.ExpireTime must be %s, not %s", expected_expire, m.ExpireTime)
	}

}

func TestJob_Run_ExecutionFailedMetadataSetted(t *testing.T) {
	now := time.Now()
	maxAge, _ := time.ParseDuration("30m")
	executor := &TestFailExecutor{}
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("cannot create temp dir:", err)
	}
	gcfg := &Config{MetadataDir: tmpdir}
	metaman := NewMetaMan(gcfg)
	defer os.RemoveAll(gcfg.MetadataDir)
	cfg := &JobConfig{Command: "utils.go", Namespace: "wow/fail", Gzip: true, MaxAge: maxAge}
	job := NewJob(
		"test_fail", cfg, "127.0.0.1:9999",
		".", executor, metaman,
	)

	err = job.Run()
	if err == nil {
		t.Fatal("error must not be nil")
	}
	m, err := metaman.View(job.TaskId)
	if err != nil {
		t.Fatal("cannot get metadata:", err)
	}
	if m.Success {
		t.Fatal("m.Success must be false")
	}
	if !m.Gzip {
		t.Fatal("m.Gzip must be true")
	}
	if m.Message != "Oops" {
		t.Fatalf("m.Message must be 'Oops' not '%s'", m.Message)
	}
	if m.JobName != "test_fail" {
		t.Fatal("m.JobName must be 'test_fail' not", m.JobName)
	}
	if m.Namespace != "wow/fail" {
		t.Fatal("m.Namespace must be 'wow/fail' not", m.Namespace)
	}
	if m.TaskId != job.TaskId {
		t.Fatalf("m.TaskId must be '%s' not '%s'", m.TaskId, job.TaskId)
	}
	if m.StartTime.Before(now) {
		t.Fatalf("m.StartTime before ", now)
	}
	if m.EndTime.Before(now) {
		t.Fatalf("m.EndTime before ", now)
	}
	expected_expire := m.StartTime.Add(maxAge)
	if !m.ExpireTime.Equal(expected_expire) {
		t.Fatalf("m.ExpireTime must be %s, not %s", expected_expire, m.ExpireTime)
	}

}

func TestJob_Run_FailedNoSuchCommand(t *testing.T) {
	executor := &TestOkExecutor{}
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("cannot create temp dir:", err)
	}
	gcfg := &Config{MetadataDir: tmpdir}
	metaman := NewMetaMan(gcfg)
	defer os.RemoveAll(gcfg.MetadataDir)
	cfg := &JobConfig{Command: "DOES_NOT_EXIST"}
	job := NewJob(
		"test_fail", cfg, "127.0.0.1:9999",
		".", executor, metaman,
	)

	err = job.Run()
	if err == nil {
		t.Fatal("error must not be nil")
	}
	m, err := metaman.View(job.TaskId)
	if err != nil {
		t.Fatal("cannot get metadata:", err)
	}

	if m.Success {
		t.Fatal("m.Success must be false")
	}
	if m.Message != "open DOES_NOT_EXIST: no such file or directory" {
		t.Fatalf("bad m.Message: %s", m.Message)
	}

}

func TestJob_Run_FailedCannotAddMetadata(t *testing.T) {
	executor := &TestOkExecutor{}
	gcfg := &Config{MetadataDir: "/DOES_NOT_EXIST"}
	metaman := NewMetaMan(gcfg)
	defer os.RemoveAll(gcfg.MetadataDir)
	cfg := &JobConfig{}
	job := NewJob(
		"test_fail", cfg, "127.0.0.1:9999",
		".", executor, metaman,
	)

	err := job.Run()
	if err.Error() == "wow" {
		t.Fatal("bad err", err)
	}
}
