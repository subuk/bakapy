package bakapy

import (
	"testing"
)

func TestJobManagerAddJobOk(t *testing.T) {
	m := NewStorageJobManager()
	m.AddJob(&StorageCurrentJob{TaskId: "test-job"})
	if _, exist := m.currentJobs["test-job"]; !exist {
		t.Fatal("test-job must be present in currentJobs")
	}
}

func TestJobManagerRemoveJobOk(t *testing.T) {
	m := NewStorageJobManager()
	m.currentJobs["test-job"] = StorageCurrentJob{TaskId: "test-job"}
	m.RemoveJob("test-job")
	if _, exist := m.GetJob("test-job"); exist {
		t.Fatal("test-job must not present in currentJobs")
	}
}

func TestJobManagerAddConnectionOk(t *testing.T) {
	m := NewStorageJobManager()
	m.AddConnection("test-job")
	m.AddConnection("test-job")
	if count := m.jobConnectionCount["test-job"]; count != 2 {
		t.Fatal("connection count must be 2, now", count)
	}
}

func TestJobManagerRemoveConnectionOk(t *testing.T) {
	m := NewStorageJobManager()
	m.jobConnectionCount["test"] = 10
	m.RemoveConnection("test")
	if count := m.jobConnectionCount["test"]; count != 9 {
		t.Fatal("connection count must be 9, now", count)
	}
}

func TestJobManagerRemoveConnectionLast(t *testing.T) {
	m := NewStorageJobManager()
	m.jobConnectionCount["test"] = 1
	m.RemoveConnection("test")
	if _, exist := m.jobConnectionCount["test"]; exist {
		t.Fatal("key 'test' mustn't exist in map jobConnectionCount")
	}
}

func TestJobManagerRemoveConnectionLastJobExist(t *testing.T) {
	m := NewStorageJobManager()
	m.jobConnectionCount["test"] = 1
	m.currentJobs["test"] = StorageCurrentJob{}
	m.RemoveConnection("test")
	if _, exist := m.jobConnectionCount["test"]; !exist {
		t.Fatal("key 'test' must exist in map jobConnectionCount")
	}
}

func TestJobManagerGetJobOk(t *testing.T) {
	m := NewStorageJobManager()
	m.currentJobs["test-1"] = StorageCurrentJob{TaskId: "test-1"}
	job, _ := m.GetJob("test-1")
	if job.TaskId != "test-1" {
		t.Fatal("job task id must be 'test-1', now", job.TaskId)
	}
}

func TestJobManagerJobConnectionCountOk(t *testing.T) {
	m := NewStorageJobManager()
	m.jobConnectionCount["test-job"] = 4
	if count := m.JobConnectionCount("test-job"); count != 4 {
		t.Fatal("connection count must be 4, now", count)
	}
}

func TestJobManagerJobConnectionCountUnknownJobOk(t *testing.T) {
	m := NewStorageJobManager()
	if count := m.JobConnectionCount("test-job-oneone"); count != 0 {
		t.Fatal("connection count must be 0, now", count)
	}
}
