package bakapy

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	ospath "path"
	"sync"
	"time"
)

type MetaManager interface {
	Keys() ([]TaskId, error)
	View(id TaskId) (Metadata, error)
	ViewAll() chan viewIterItem
	Add(jobName, namespace, command string, taskId TaskId, gzip bool, maxAge time.Duration) error
	Update(id TaskId, up func(m *Metadata)) error
	Remove(id TaskId) error
}

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

func (m *MetaMan) get(id TaskId) (*Metadata, error) {
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

func (m *MetaMan) View(id TaskId) (Metadata, error) {
	md, err := m.get(id)
	if err != nil {
		return Metadata{}, err
	}
	return *md, err
}

func (m *MetaMan) getForUpdate(id TaskId) (*Metadata, error) {
	lock, exist := m.taken[id]
	if !exist {
		lock = new(sync.Mutex)
		m.taken[id] = lock
	}
	lock.Lock()
	data, err := m.get(id)
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
			metadata, err := m.get(key)
			ch <- viewIterItem{metadata, err}
		}
		close(ch)
	}()
	return ch
}

func (m *MetaMan) save(id TaskId, metadata *Metadata) error {

	saveTo := ospath.Join(m.RootDir, id.String())
	saveToTmp := saveTo + ".inpr"

	err := os.MkdirAll(ospath.Dir(saveTo), 0750)
	if err != nil {
		return err
	}
	file, err := os.Create(saveToTmp)
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

	err = os.Rename(saveToTmp, saveTo)
	if err != nil {
		return err
	}
	if lock, exist := m.taken[id]; exist {
		lock.Unlock()
	}
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
	return m.save(taskId, metadata)
}

func (m *MetaMan) Update(id TaskId, up func(m *Metadata)) error {
	md, err := m.getForUpdate(id)
	if err != nil {
		return err
	}
	up(md)
	err = m.save(id, md)
	if err != nil {
		panic(err)
	}
	return nil
}

func (m *MetaMan) Remove(id TaskId) error {
	return os.Remove(ospath.Join(m.RootDir, id.String()))
}