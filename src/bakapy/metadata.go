package bakapy

import (
	"fmt"
	"time"
)

type MetadataFileEntry struct {
	Name       string
	Size       int64
	SourceAddr string
	StartTime  time.Time
	EndTime    time.Time
}

func (m *MetadataFileEntry) String() string {
	return fmt.Sprintf(`{name: "%s", size: "%d", start_time: "%s", end_time: "%s"`,
		m.Name, m.Size, m.StartTime, m.EndTime)
}

type Metadata struct {
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
	Files      []MetadataFileEntry
	Pid        int
	RetCode    uint
	Script     []byte
	Output     []byte
	Errput     []byte
	Config     JobConfig
}

func (metadata *Metadata) Duration() time.Duration {
	if (metadata.EndTime == time.Time{}) || (metadata.StartTime == time.Time{}) {
		return time.Duration(0)
	}
	if metadata.StartTime.After(metadata.EndTime) {
		return time.Duration(0)
	}
	return metadata.EndTime.Sub(metadata.StartTime)
}

func (metadata *Metadata) AvgSpeed() int64 {
	if int64(metadata.Duration().Seconds()) == 0 {
		return 0
	}
	return metadata.TotalSize / int64(metadata.Duration().Seconds())
}
