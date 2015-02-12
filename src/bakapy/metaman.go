package bakapy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	ospath "path"
	"sync"
	"time"
)

type MetaMan struct {
	RootDir string
	taken   map[TaskId]*sync.Mutex
}

func NewMetaMan(cfg *Config) *MetaMan {
	return &MetaMan{
		RootDir: cfg.MetadataDir,
		taken:   make(map[TaskId]*sync.Mutex),
	}
}

func (m *MetaMan) Keys() ([]TaskId, error) {
	dir, err := ioutil.ReadDir(m.RootDir)
	if err != nil {
		return nil, err
	}
	var ret []TaskId
	for _, f := range dir {
		ret = append(ret, TaskId(f.Name()))
	}
	return ret, nil
}

func (m *MetaMan) Get(id TaskId) (*Metadata, error) {
	data, err := ioutil.ReadFile(ospath.Join(m.RootDir, id.String()))
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{}
	err = json.Unmarshal(data, metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func (m *MetaMan) GetForUpdate(id TaskId) (*Metadata, error) {
	lock, exist := m.taken[id]
	if !exist {
		lock := &sync.Mutex{}
		m.taken[id] = lock
	}
	lock.Lock()
	data, err := m.Get(id)
	if err != nil {
		lock.Unlock()
		return nil, err
	}
	return data, nil
}

type viewIterItem struct {
	metadata *Metadata
	err      error
}

func (m *MetaMan) ViewAll() chan viewIterItem {
	ch := make(chan viewIterItem)
	go func() {
		paths, err := m.Keys()
		if err != nil {
			ch <- viewIterItem{nil, err}
			close(ch)
			return
		}
		for _, key := range paths {
			metadata, err := m.Get(key)
			ch <- viewIterItem{metadata, err}
		}
		close(ch)
	}()
	return ch
}

func (m *MetaMan) Commit(id TaskId, metadata *Metadata) error {
	lock, exist := m.taken[id]
	if !exist {
		return fmt.Errorf("metadata %s not taken", id)
	}

	saveTo := ospath.Join(m.RootDir, id.String())

	err := os.MkdirAll(ospath.Dir(saveTo), 0750)
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

	lock.Unlock()
	return nil
}

func (m *MetaMan) Add(jobName, namespace, command string, taskId TaskId, gzip bool, maxAge time.Duration) error {
	now := time.Now().UTC()
	metadata := &Metadata{
		TaskId:     taskId,
		JobName:    jobName,
		Gzip:       gzip,
		Namespace:  namespace,
		Command:    command,
		StartTime:  now,
		ExpireTime: now.Add(maxAge),
	}
	return m.Commit(taskId, metadata)
}

func (m *MetaMan) Update(id TaskId, up func(m *Metadata)) error {
	md, err := m.GetForUpdate(id)
	if err != nil {
		return err
	}
	up(md)
	return m.Commit(id, md)
}

func (m *MetaMan) Remove(id TaskId) error {
	md, err := m.GetForUpdate(id)
	if err != nil {
		return err
	}
	return os.Remove(ospath.Join(m.RootDir, md.TaskId.String()))
}
