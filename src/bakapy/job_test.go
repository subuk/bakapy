package bakapy

import (
	"errors"
	"io"
	"testing"
	"time"
)

type TestJober struct{}

func (j *TestJober) AddJob(currentJob *StorageCurrentJob) {}
func (j *TestJober) RemoveJob(id TaskId)                  {}
func (j *TestJober) WaitJob(taskId TaskId)                {}

type TestJoberPushFile struct {
	TestJober
	ch chan MetadataFileEntry
}

func (j *TestJoberPushFile) AddJob(currentJob *StorageCurrentJob) {
	j.ch = currentJob.FileAddChan
	f1 := MetadataFileEntry{
		Name:       currentJob.Namespace + "/" + "wow.txt",
		Size:       1234,
		SourceAddr: "1.1.1.1:1234",
		StartTime:  time.Date(2001, 10, 4, 5, 0, 0, 0, time.UTC),
		EndTime:    time.Date(2001, 10, 4, 6, 0, 0, 0, time.UTC),
	}
	f2 := MetadataFileEntry{
		Name:       currentJob.Namespace + "/" + "hello.txt",
		Size:       12345,
		SourceAddr: "1.1.1.1:12345",
		StartTime:  time.Date(2001, 10, 4, 7, 0, 0, 0, time.UTC),
		EndTime:    time.Date(2001, 10, 4, 8, 0, 0, 0, time.UTC),
	}

	currentJob.FileAddChan <- f1
	currentJob.FileAddChan <- f2
}

func (j *TestJoberPushFile) WaitJob(taskId TaskId) {
	for {
		if len(j.ch) == 0 {
			break
		}
		time.Sleep(time.Nanosecond)
	}
}

type TestOkExecutor struct{}

func (e *TestOkExecutor) Execute(script []byte, output io.Writer, errput io.Writer) error {
	return nil
}

type TestFailExecutor struct{}

func (e *TestFailExecutor) Execute(script []byte, output io.Writer, errput io.Writer) error {
	return errors.New("Oops")
}

func TestJob_Run_MetadataFieldSetted(t *testing.T) {
	executor := &TestOkExecutor{}
	jober := &TestJober{}

	cfg := &JobConfig{Command: "utils.go"}
	job := NewJob(
		"test", cfg, "127.0.0.1:9999",
		".", jober, executor,
	)

	m := job.Run()

	if !m.Success {
		t.Fatal("m.Success must be true")
	}
}

func TestJob_Run_ExecutionOkMetadataSetted(t *testing.T) {
	now := time.Now()
	executor := &TestOkExecutor{}
	jober := &TestJober{}
	maxAge, _ := time.ParseDuration("30m")
	cfg := &JobConfig{Command: "utils.go", Namespace: "wow", Gzip: true, MaxAge: maxAge}
	job := NewJob(
		"test", cfg, "127.0.0.1:9999",
		".", jober, executor,
	)

	m := job.Run()

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
	jober := &TestJober{}

	cfg := &JobConfig{Command: "utils.go", Namespace: "wow/fail", Gzip: false, MaxAge: maxAge}
	job := NewJob(
		"test_fail", cfg, "127.0.0.1:9999",
		".", jober, executor,
	)

	m := job.Run()

	if m.Success {
		t.Fatal("m.Success must be false")
	}
	if m.Gzip {
		t.Fatal("m.Success must be false")
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
	jober := &TestJober{}

	cfg := &JobConfig{Command: "DOES_NOT_EXIST"}
	job := NewJob(
		"test_fail", cfg, "127.0.0.1:9999",
		".", jober, executor,
	)

	m := job.Run()

	if m.Success {
		t.Fatal("m.Success must be false")
	}
	if m.Message != "open DOES_NOT_EXIST: no such file or directory" {
		t.Fatalf("m.Message must be 'open DOES_NOT_EXIST: no such file or directory' not '%s'", m.Message)
	}

}

func TestJob_Run_MetadataFileEntrysAdded(t *testing.T) {
	executor := &TestOkExecutor{}
	jober := &TestJoberPushFile{}

	cfg := &JobConfig{Command: "utils.go", Namespace: "wow/fail"}
	job := NewJob(
		"test_fail", cfg, "127.0.0.1:9999",
		".", jober, executor,
	)

	m := job.Run()

	if !m.Success {
		t.Fatal("m.Success must be true. Message", m.Message)
	}
	if len(m.Files) != 2 {
		t.Fatal("m.Files length must be 2 not", len(m.Files))
	}

	f1 := m.Files[0]
	if f1.Name != "wow/fail/wow.txt" {
		t.Fatal("f1.Name must be 'wow/fail/wow.txt' not", f1.Name)
	}
	if f1.Size != 1234 {
		t.Fatal("f1.Size must be 1234 not", f1.Size)
	}
	if f1.SourceAddr != "1.1.1.1:1234" {
		t.Fatal("f1.SourceAddr must be 1.1.1.1:1234 not", f1.SourceAddr)
	}
	if f1.StartTime.String() != "2001-10-04 05:00:00 +0000 UTC" {
		t.Fatal("f1.StartTime must be '2001-10-04 05:00:00 +0000 UTC' not", f1.StartTime)
	}
	if f1.EndTime.String() != "2001-10-04 06:00:00 +0000 UTC" {
		t.Fatal("f1.EndTime must be '2001-10-04 06:00:00 +0000 UTC' not", f1.EndTime)
	}

	f2 := m.Files[1]
	if f2.Name != "wow/fail/hello.txt" {
		t.Fatal("f2.Name must be 'wow/fail/hello.txt' not", f2.Name)
	}
	if f2.Size != 12345 {
		t.Fatal("f2.Size must be 12345 not", f2.Size)
	}
	if f2.SourceAddr != "1.1.1.1:12345" {
		t.Fatal("f2.SourceAddr must be 1.1.1.1:12345 not", f2.SourceAddr)
	}
	if f2.StartTime.String() != "2001-10-04 07:00:00 +0000 UTC" {
		t.Fatal("f2.StartTime must be '2001-10-04 07:00:00 +0000 UTC' not", f2.StartTime)
	}
	if f2.EndTime.String() != "2001-10-04 08:00:00 +0000 UTC" {
		t.Fatal("f2.EndTime must be '2001-10-04 08:00:00 +0000 UTC' not", f2.EndTime)
	}

}
