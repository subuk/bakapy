package bakapy

import (
	"os"
	"path"
	"time"
)

func (stor *Storage) CleanupExpired() error {
	for taskId := range stor.metaman.Keys() {
		metadata, err := stor.metaman.View(taskId)
		if err != nil {
			stor.logger.Warning("error reading metadata file: %s", err.Error())
			continue
		}
		if metadata.ExpireTime.After(time.Now()) {
			continue
		}

		stor.logger.Info("removing files for expired task %s(%s)",
			metadata.JobName, metadata.TaskId)

		removeErrs := false
		for _, fileMeta := range metadata.Files {
			absPath := path.Join(stor.RootDir, metadata.Namespace, fileMeta.Name)
			stor.logger.Info("removing file %s", absPath)
			err = os.Remove(absPath)
			if err != nil {
				removeErrs = true
				stor.logger.Warning("cannot remove file %s: %s", absPath, err.Error())
			}
		}
		if !removeErrs {
			stor.logger.Info("removing metadata %s", metadata.TaskId)
			err := stor.metaman.Remove(metadata.TaskId)
			if err != nil {
				stor.logger.Warning("cannot metadata %s: %s", metadata.TaskId, err.Error())
				continue
			}
		}
	}
	return nil
}
