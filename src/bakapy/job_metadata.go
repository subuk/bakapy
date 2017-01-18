package bakapy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type JobMetadataFile struct {
	Name       string
	Size       int64
	SourceAddr string
	StartTime  time.Time
	EndTime    time.Time
}

func (m *JobMetadataFile) String() string {
	return fmt.Sprintf(`{name: "%s", size: "%d", start_time: "%s", end_time: "%s"`,
		m.Name, m.Size, m.StartTime, m.EndTime)
}

type MetadataSortByStartTime []JobMetadata

func (slice MetadataSortByStartTime) Len() int {
	return len(slice)
}

func (slice MetadataSortByStartTime) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice MetadataSortByStartTime) Less(i, j int) bool {
	return slice[i].StartTime.Unix() < slice[j].StartTime.Unix()
}

type JobMetadata struct {
	JobName    string
	Gzip       bool
	Namespace  string
	TaskId     TaskId
	Command    string
	Success    bool
	Message    string
	TotalSize  int64
	StartTime  time.Time
	EndTime    time.Time
	ExpireTime time.Time
	Files      []JobMetadataFile
	Pid        int
	RetCode    uint
	Script     []byte
	Output     []byte
	Errput     []byte
	Config     JobConfig
	Corrupted  bool   `json:"-"`
	Filepath   string `json:"-"`
}

func (metadata *JobMetadata) Duration() time.Duration {
	if (metadata.EndTime == time.Time{}) || (metadata.StartTime == time.Time{}) {
		return time.Duration(0)
	}
	if metadata.StartTime.After(metadata.EndTime) {
		return time.Duration(0)
	}
	return metadata.EndTime.Sub(metadata.StartTime)
}

func (metadata *JobMetadata) AvgSpeed() int64 {
	if int64(metadata.Duration().Seconds()) == 0 {
		return 0
	}
	return metadata.TotalSize / int64(metadata.Duration().Seconds())
}

func (metadata *JobMetadata) Save(saveTo string) error {
	err := os.MkdirAll(path.Dir(saveTo), 0750)
	if err != nil {
		return err
	}
	file, err := os.Create(saveTo)
	if err != nil {
		return err
	}
	defer file.Close()
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	return nil
}

func LoadJobMetadata(path string) (*JobMetadata, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	metadata := JobMetadata{}
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}
