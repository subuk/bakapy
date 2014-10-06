package bakapy

import (
	"github.com/op/go-logging"
	"sync"
)

type getJobRequest struct {
	id   TaskId
	resp chan StorageCurrentJob
}

type StorageJobManager struct {
	jobMu              sync.RWMutex
	connMu             sync.RWMutex
	currentJobs        map[TaskId]StorageCurrentJob
	jobConnectionCount map[TaskId]int
	logger             *logging.Logger
}

func NewStorageJobManager() *StorageJobManager {
	m := &StorageJobManager{
		currentJobs:        make(map[TaskId]StorageCurrentJob, 30),
		jobConnectionCount: make(map[TaskId]int, 30),
		logger:             logging.MustGetLogger("bakapy.storage.jobmanager"),
	}
	return m
}

func (m *StorageJobManager) JobConnectionCount(taskId TaskId) int {
	m.connMu.RLock()
	defer m.connMu.RUnlock()
	count, exist := m.jobConnectionCount[taskId]
	if !exist {
		return 0
	}
	return count
}

func (m *StorageJobManager) AddJob(job *StorageCurrentJob) {
	m.jobMu.Lock()
	defer m.jobMu.Unlock()
	m.currentJobs[job.TaskId] = *job
}

func (m *StorageJobManager) RemoveJob(id TaskId) {
	m.jobMu.Lock()
	defer m.jobMu.Unlock()
	delete(m.currentJobs, id)
}

func (m *StorageJobManager) AddConnection(id TaskId) {
	m.connMu.Lock()
	defer m.connMu.Unlock()
	_, exist := m.jobConnectionCount[id]
	if !exist {
		m.jobConnectionCount[id] = 0
	}
	m.jobConnectionCount[id] += 1
	m.logger.Debug("connection count for task %s increased, now %d",
		id, m.jobConnectionCount[id])
}

func (m *StorageJobManager) RemoveConnection(id TaskId) {
	m.connMu.Lock()
	defer m.connMu.Unlock()

	_, exist := m.jobConnectionCount[id]
	if !exist {
		return
	}
	m.jobConnectionCount[id] -= 1
	m.logger.Debug("connection count for task %s decreased, now %d",
		id, m.jobConnectionCount[id])

	_, exist = m.GetJob(id)
	if !exist && m.jobConnectionCount[id] <= 0 {
		delete(m.jobConnectionCount, id)
	}

}

func (m *StorageJobManager) GetJob(id TaskId) (StorageCurrentJob, bool) {
	m.jobMu.RLock()
	defer m.jobMu.RUnlock()
	job, exist := m.currentJobs[id]
	return job, exist
}
