package bakapy

import (
	"github.com/op/go-logging"
)

type StorageJobManager struct {
	AddJob           chan StorageCurrentJob
	AddConnection    chan TaskId
	RemoveJob        chan StorageCurrentJob
	RemoveConnection chan TaskId

	currentJobs        map[TaskId]StorageCurrentJob
	jobConnectionCount map[TaskId]int
}

func NewStorageJobManager() *StorageJobManager {
	m := &StorageJobManager{
		AddJob:             make(chan StorageCurrentJob),
		AddConnection:      make(chan TaskId),
		RemoveJob:          make(chan StorageCurrentJob),
		RemoveConnection:   make(chan TaskId),
		currentJobs:        make(map[TaskId]StorageCurrentJob, 30),
		jobConnectionCount: make(map[TaskId]int, 30),
	}
	go m.handle()

	return m
}

func (m *StorageJobManager) GetJobs() map[TaskId]StorageCurrentJob {
	return m.currentJobs
}

func (m *StorageJobManager) GetJob(taskId TaskId) *StorageCurrentJob {
	job, exist := m.currentJobs[taskId]
	if !exist {
		return nil
	}
	return &job
}

func (m *StorageJobManager) handle() {
	logger := logging.MustGetLogger("bakapy.storage.jobmanager")
	for {
		select {

		case activeJob := <-m.AddJob:
			logger.Debug("adding job %s", activeJob.TaskId)
			m.currentJobs[activeJob.TaskId] = activeJob

		case activeJob := <-m.RemoveJob:
			logger.Debug("removing job %s", activeJob.TaskId)
			delete(m.currentJobs, activeJob.TaskId)

		case taskId := <-m.AddConnection:
			_, exist := m.jobConnectionCount[taskId]
			if !exist {
				m.jobConnectionCount[taskId] = 0
			}
			m.jobConnectionCount[taskId] += 1
			logger.Debug("connection count for task %s increased, now %d",
				taskId, m.jobConnectionCount[taskId])

		case taskId := <-m.RemoveConnection:
			_, exist := m.jobConnectionCount[taskId]
			if !exist {
				m.jobConnectionCount[taskId] = 0
			}
			m.jobConnectionCount[taskId] -= 1
			logger.Debug("connection count for task %s decreased, now %d",
				taskId, m.jobConnectionCount[taskId])

		}

	}
}
