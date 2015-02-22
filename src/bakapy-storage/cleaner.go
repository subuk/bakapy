package main

import (
	"bakapy"
	"github.com/op/go-logging"
	"os"
	"time"
)

type Cleaner func(metaman bakapy.MetaManager, storage Storage) error

func CleanupExpiredJobs(metaman bakapy.MetaManager, storage Storage) error {
	logger := logging.MustGetLogger("bakapy.cleaner.CleanupExpiredFiles")
	logger.Debug("cleaning up storage")
	for taskId := range metaman.Keys() {
		md, err := metaman.View(taskId)
		if err != nil {
			logger.Warning("error reading metadata %s: %s", taskId, err.Error())
			continue
		}
		if md.ExpireTime.After(time.Now()) {
			continue
		}

		logger.Info("removing files for expired task %s(%s)",
			md.JobName, md.TaskId)

		removeErrs := false
		for _, fileMeta := range md.Files {
			logger.Info("removing file {%s}%s", md.Namespace, fileMeta.Name)
			err = storage.Remove(md.Namespace, fileMeta.Name)
			if err != nil {
				if os.IsNotExist(err) {
					logger.Warning("file {%s}%s does not exist, skipping", md.Namespace, fileMeta.Name)
					continue
				}
				removeErrs = true
				logger.Warning("cannot remove file {%s}%s: %s", md.Namespace, fileMeta.Name, err.Error())
			}
		}
		if !removeErrs {
			logger.Info("removing metadata %s", md.TaskId)
			err := metaman.Remove(md.TaskId)
			if err != nil {
				logger.Warning("cannot remove metadata %s: %s", md.TaskId, err.Error())
				continue
			}
		}
	}
	return nil
}
