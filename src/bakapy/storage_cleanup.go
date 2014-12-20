package bakapy

import (
	"os"
	"path"
	"path/filepath"
	"time"
)

func (stor *Storage) CleanupExpired() error {
	visit := func(metaPath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() {
			return nil
		}

		metadata, err := LoadJobMetadata(metaPath)
		if err != nil {
			stor.logger.Warning("corrupt metadata file %s: %s", metaPath, err.Error())
			return nil
		}
		if metadata.ExpireTime.After(time.Now()) {
			return nil
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
			stor.logger.Info("removing metadata %s", metaPath)
			err = os.Remove(metaPath)
			if err != nil {
				stor.logger.Warning("cannot remove file %s: %s", metaPath, err.Error())
				return nil
			}
		}
		return nil
	}

	return filepath.Walk(stor.MetadataDir, visit)
}
