package bakapy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"io/ioutil"
	"os"
	ospath "path"
	"sync"
)

type MetaManager interface {
	Keys() chan TaskId
	View(id TaskId) (Metadata, error)
	Add(id TaskId, md Metadata) error
	Update(id TaskId, up func(m *Metadata)) error
	Remove(id TaskId) error
}

type MetaMan struct {
	sync.Mutex
	RootDir string
	taken   map[TaskId]*sync.Mutex
	logger  *logging.Logger
}

func NewMetaMan(cfg *Config) *MetaMan {
	return &MetaMan{
		RootDir: cfg.MetadataDir,
		taken:   make(map[TaskId]*sync.Mutex),
		logger:  logging.MustGetLogger("bakapy.metaman"),
	}
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

func (m *MetaMan) getForUpdate(id TaskId) (*Metadata, error) {
	m.logger.Debug("getting for update metadata for task id %s", id)
	lock, exist := m.taken[id]
	if !exist {
		lock = new(sync.Mutex)
		m.Lock()
		m.taken[id] = lock
		m.Unlock()
	}
	lock.Lock()
	data, err := m.get(id)
	if err != nil {
		lock.Unlock()
		m.Lock()
		delete(m.taken, id)
		m.Unlock()
		return nil, err
	}
	return data, nil
}

type viewIterItem struct {
	metadata *Metadata
	err      error
}

func (m *MetaMan) save(id TaskId, metadata *Metadata) error {
	saveTo := ospath.Join(m.RootDir, id.String())
	saveToTmp := saveTo + ".inpr"
	m.logger.Debug("saving metadata for task id %s to %s", id, saveTo)

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
		m.Lock()
		delete(m.taken, id)
		m.Unlock()
	}
	return nil
}

func (m *MetaMan) Keys() chan TaskId {
	dir, err := ioutil.ReadDir(m.RootDir)
	if err != nil {
		panic(fmt.Errorf("cannot list metadata directory: %s", err))
	}
	ch := make(chan TaskId, 100)
	go func() {
		for _, f := range dir {
			ch <- TaskId(f.Name())
		}
		close(ch)
	}()
	return ch
}

func (m *MetaMan) View(id TaskId) (Metadata, error) {
	md, err := m.get(id)
	if err != nil {
		return Metadata{}, err
	}
	return *md, err
}

func (m *MetaMan) Add(id TaskId, md Metadata) error {
	m.logger.Debug("adding metadata for task id %s", id)
	md.TaskId = id
	if _, err := m.View(id); err == nil {
		return fmt.Errorf("metadata for task %s already exist", id)
	}
	return m.save(id, &md)
}

func (m *MetaMan) Update(id TaskId, up func(m *Metadata)) error {
	md, err := m.getForUpdate(id)
	if err != nil {
		return err
	}
	up(md)
	return m.save(id, md)
}

func (m *MetaMan) Remove(id TaskId) error {
	return os.Remove(ospath.Join(m.RootDir, id.String()))
}
