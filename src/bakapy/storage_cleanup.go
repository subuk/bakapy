package bakapy

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

func (stor *Storage) CleanupExpired() error {
	jobMetadataList := map[string][]JobMetadata{}
	corrupted := []string{}
	corruptedDir := stor.MetadataDir + "_corrupted"
	if err := os.MkdirAll(corruptedDir, 0755); err != nil {
		return err
	}

	visit := func(metaPath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		metadata, err := LoadJobMetadata(metaPath)
		if err != nil {
			corrupted = append(corrupted, metaPath)
			return nil
		}
		metadata.Filepath = metaPath
		jobMetadataList[metadata.JobName] = append(jobMetadataList[metadata.JobName], *metadata)
		return nil
	}

	if err := filepath.Walk(stor.MetadataDir, visit); err != nil {
		return err
	}

	for _, metadataPath := range corrupted {
		_, oldFilename := path.Split(metadataPath)
		newFullPath := path.Join(corruptedDir, oldFilename)
		stor.logger.Warning("moving corrupted metadata file %s to %s", metadataPath, newFullPath)
		if err := os.Rename(metadataPath, newFullPath); err != nil {
			stor.logger.Warning("cannot move corrupted metadata file: %s", err)
		}
	}

	for jobName, jobMetadatas := range jobMetadataList {
		sort.Sort(MetadataSortByStartTime(jobMetadatas))

		if !jobMetadatas[len(jobMetadatas)-1].Success {
			stor.logger.Warning("skipping cleanup for job %s due to last task failure", jobName)
			continue
		}

		for _, metadata := range jobMetadatas {
			if time.Now().Before(metadata.ExpireTime) {
				continue
			}
			for _, fileMeta := range metadata.Files {
				dataFilePath := path.Join(stor.RootDir, metadata.Namespace, fileMeta.Name)
				stor.logger.Info("removing file %s", dataFilePath)
				if err := os.Remove(dataFilePath); err != nil {
					stor.logger.Warning("failed to remove file %s: %s", dataFilePath, err)
				}
			}
			if err := os.Remove(metadata.Filepath); err != nil {
				stor.logger.Warning("failed to remove metadata file: %s", err)
			}
		}
	}
	return nil
}
