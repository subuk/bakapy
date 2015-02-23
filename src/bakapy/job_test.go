package bakapy

import (
	"errors"
	"io"
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

type TestCustomExecutor struct {
	execute func(script []byte, output io.Writer, errput io.Writer) error
}

func (e *TestCustomExecutor) Execute(script []byte, output io.Writer, errput io.Writer) error {
	return e.execute(script, output, errput)
}

type TestNotificator struct {
	calledMd    *Metadata
	calledMdErr error
}

func (t *TestNotificator) JobFinished(md Metadata) error {
	t.calledMd = &md
	return nil
}

func (t *TestNotificator) MetadataAccessFailed(err error) error {
	t.calledMdErr = err
	return nil
}

func (t *TestNotificator) Name() string {
	return "test-notificator"
}

type TestMetaMan struct {
	stor   map[TaskId]Metadata
	addErr error
}

func NewTestMockMetaMan() *TestMetaMan {
	return &TestMetaMan{
		stor: make(map[TaskId]Metadata),
	}
}

func (mm *TestMetaMan) Keys() chan TaskId {
	ch := make(chan TaskId)
	go func() {
		for key, _ := range mm.stor {
			ch <- key
		}
		close(ch)
	}()
	return ch
}

func (mm *TestMetaMan) View(id TaskId) (Metadata, error) {
	md, ok := mm.stor[id]
	if !ok {
		return Metadata{}, errors.New("does not exist")
	}
	return md, nil
}

func (mm *TestMetaMan) Add(id TaskId, md Metadata) error {
	md.TaskId = id
	if mm.addErr == nil {
		mm.stor[id] = md
		return nil
	}
	return mm.addErr
}

func (mm *TestMetaMan) Update(id TaskId, up func(*Metadata)) error {
	md, err := mm.View(id)
	if err != nil {
		return err
	}
	up(&md)
	mm.stor[id] = md
	return nil
}

func (mm *TestMetaMan) Remove(id TaskId) error {
	delete(mm.stor, id)
	return nil
}

func (mm *TestMetaMan) AddFile(id TaskId, fm MetadataFileEntry) error {
	md := mm.stor[id]
	md.Files = append(md.Files, fm)
	mm.stor[id] = md
	return nil
}

func TestJob_Run_ExecutionOkMetadataSetted(t *testing.T) {
	now := time.Now()
	executor := &TestOkExecutor{}
	maxAge, _ := time.ParseDuration("30m")
	cfg := &JobConfig{Command: "utils.go", Namespace: "wow", Gzip: true, MaxAge: maxAge}
	metaman := NewTestMockMetaMan()
	spool := &TestScriptPool{nil, nil, ""}
	notify := &TestNotificator{}
	job := NewJob(
		"test", cfg, "127.0.0.1:9999",
		spool, executor, metaman, notify,
	)

	job.Run()
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
	metaman := NewTestMockMetaMan()
	cfg := &JobConfig{Command: "utils.go", Namespace: "wow/fail", Gzip: true, MaxAge: maxAge}
	spool := &TestScriptPool{nil, nil, ""}
	notify := &TestNotificator{}
	job := NewJob(
		"test_fail", cfg, "127.0.0.1:9999",
		spool, executor, metaman, notify,
	)

	job.Run()
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

func TestJob_Run_FailedCannotAddMetadata(t *testing.T) {
	executor := &TestOkExecutor{}
	metaman := NewTestMockMetaMan()
	spool := &TestScriptPool{nil, nil, ""}
	cfg := &JobConfig{}
	notify := &TestNotificator{}
	job := NewJob(
		"test_fail", cfg, "127.0.0.1:9999",
		spool, executor, metaman, notify,
	)
	metaman.addErr = errors.New("test error wow are")
	job.Run()
	if err := notify.calledMdErr; err.Error() != "cannot add metadata: test error wow are" {
		t.Fatal("bad err", err)
	}
}

func TestJob_Run_FailedCannotGetScript(t *testing.T) {
	executor := &TestOkExecutor{}
	metaman := NewTestMockMetaMan()
	spool := &TestScriptPool{errors.New("test bad script"), nil, ""}
	cfg := &JobConfig{Command: "wowcmd"}
	notify := &TestNotificator{}
	job := NewJob(
		"test_fail", cfg, "127.0.0.1:9999",
		spool, executor, metaman, notify,
	)

	job.Run()
	md, err := metaman.View(job.TaskId)
	if err != nil {
		t.Fatal("unexpected err", err)
	}
	if md.Message != "cannot get job script: cannot find backup script wowcmd: test bad script" {
		t.Fatal("bad err", md.Message)
	}
}

func TestJob_Run_ExecutionOkMetadataTotalSizeCalculated(t *testing.T) {
	cfg := &JobConfig{}

	metaman := NewTestMockMetaMan()
	spool := &TestScriptPool{nil, nil, ""}

	releaseExecute := make(chan int)
	executor := &TestCustomExecutor{
		execute: func(script []byte, output io.Writer, errput io.Writer) error {
			<-releaseExecute
			return nil
		},
	}
	notify := &TestNotificator{}
	job := NewJob(
		"test", cfg, "127.0.0.1:9999",
		spool, executor, metaman, notify,
	)

	waitJobRun := make(chan int)
	go func() {
		defer close(waitJobRun)
		job.Run()
		md, err := metaman.View(job.TaskId)
		if err != nil {
			t.Fatal("unexpected error", err)
			return
		}
		if !md.Success {
			t.Fatal("md.Success == true expected")
		}

		if md.TotalSize != 300 {
			t.Fatal("Metadata total size must be 300, not", md.TotalSize)
			return
		}
	}()

	go metaman.Update(job.TaskId, func(md *Metadata) {
		md.Files = append(md.Files, MetadataFileEntry{
			Name: "test1.txt",
			Size: 100,
		})
		md.Files = append(md.Files, MetadataFileEntry{
			Name: "test1.txt",
			Size: 150,
		})
		md.Files = append(md.Files, MetadataFileEntry{
			Name: "test1.txt",
			Size: 50,
		})
		close(releaseExecute)
	})
	<-waitJobRun
}

func TestJob_Run_ExecutionFailedMetadataTotalSizeCalculated(t *testing.T) {
	cfg := &JobConfig{}

	metaman := NewTestMockMetaMan()
	spool := &TestScriptPool{nil, nil, ""}

	releaseExecute := make(chan int)
	executor := &TestCustomExecutor{
		execute: func(script []byte, output io.Writer, errput io.Writer) error {
			<-releaseExecute
			return errors.New("test error")
		},
	}
	notify := &TestNotificator{}
	job := NewJob(
		"test", cfg, "127.0.0.1:9999",
		spool, executor, metaman, notify,
	)

	waitJobRun := make(chan int)
	go func() {
		defer close(waitJobRun)
		job.Run()
		md, err := metaman.View(job.TaskId)
		if err != nil {
			t.Fatal("unexpected error", err)
			return
		}
		if md.Success {
			t.Fatal("md.Success == false expected")
		}
		if md.TotalSize != 330 {
			t.Fatal("Metadata total size must be 330, not", md.TotalSize)
			return
		}
	}()

	go metaman.Update(job.TaskId, func(md *Metadata) {
		md.Files = append(md.Files, MetadataFileEntry{
			Name: "test1.txt",
			Size: 100,
		})
		md.Files = append(md.Files, MetadataFileEntry{
			Name: "test1.txt",
			Size: 150,
		})
		md.Files = append(md.Files, MetadataFileEntry{
			Name: "test1.txt",
			Size: 80,
		})
		close(releaseExecute)
	})
	<-waitJobRun
}
