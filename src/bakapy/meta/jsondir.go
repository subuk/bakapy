package meta

import (
	"bakapy"
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

type JSONDir struct {
	sync.Mutex
	RootDir string
	taken   map[bakapy.TaskId]*sync.Mutex
	logger  *logging.Logger
}

func NewJSONDir(root string) *JSONDir {
	return &JSONDir{
		RootDir: root,
		taken:   make(map[bakapy.TaskId]*sync.Mutex),
		logger:  logging.MustGetLogger("bakapy.JSONDir"),
	}
}

func (m *JSONDir) lockId(id bakapy.TaskId) {
	m.Lock()
	lock, exist := m.taken[id]
	if !exist {
		lock = &sync.Mutex{}
		m.taken[id] = lock
	}
	m.Unlock()
	lock.Lock()
}

func (m *JSONDir) unLockId(id bakapy.TaskId) {
	m.Lock()
	lock, exist := m.taken[id]
	if !exist {
		panic(fmt.Errorf("id %s not locked", id))
	}
	lock.Unlock()
	m.Unlock()
}

func (m *JSONDir) get(id bakapy.TaskId) (*bakapy.Metadata, error) {
	filePath := ospath.Join(m.RootDir, id.String())
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	metadata := &bakapy.Metadata{}
	err = json.Unmarshal(data, metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func (m *JSONDir) GetForUpdate(id bakapy.TaskId) (*bakapy.Metadata, error) {
	m.logger.Debug("getting for update metadata for task id %s", id)

	m.lockId(id)

	data, err := m.get(id)
	if err != nil {
		m.unLockId(id)
		return nil, err
	}
	return data, nil
}

func (m *JSONDir) Save(id bakapy.TaskId, metadata *bakapy.Metadata) error {
	defer m.unLockId(id)
	saveTo := ospath.Join(m.RootDir, id.String())
	saveToTmp := saveTo + ".inpr"
	m.logger.Debug("saving metadata for task id %s to %s", id, saveTo)
	m.logger.Debug("%s", metadata)

	err := os.MkdirAll(ospath.Dir(saveTo), 0750)
	if err != nil {
		return err
	}
	file, err := os.Create(saveToTmp)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		file.Close()
		return err
	}

	_, err = io.Copy(file, bytes.NewReader(jsonData))
	if err != nil {
		file.Close()
		return err
	}
	file.Close()
	err = os.Rename(saveToTmp, saveTo)
	if err != nil {
		return err
	}

	return nil
}

func (m *JSONDir) Keys() chan bakapy.TaskId {
	dir, err := ioutil.ReadDir(m.RootDir)
	if err != nil {
		panic(fmt.Errorf("cannot list metadata directory: %s", err))
	}
	ch := make(chan bakapy.TaskId, 100)
	go func() {
		for _, f := range dir {
			ch <- bakapy.TaskId(f.Name())
		}
		close(ch)
	}()
	return ch
}

func (m *JSONDir) View(id bakapy.TaskId) (bakapy.Metadata, error) {
	md, err := m.get(id)
	if err != nil {
		return bakapy.Metadata{}, err
	}
	return *md, err
}

func (m *JSONDir) Add(id bakapy.TaskId, md bakapy.Metadata) error {
	m.logger.Debug("adding metadata for task id %s", id)
	m.lockId(id)
	md.TaskId = id
	if _, err := m.View(id); err == nil {
		return fmt.Errorf("metadata for task %s already exist", id)
	}
	return m.Save(id, &md)
}

func (m *JSONDir) Remove(id bakapy.TaskId) error {
	m.logger.Debug("removing metadata for task id %s", id)
	m.lockId(id)
	defer m.unLockId(id)
	return os.Remove(ospath.Join(m.RootDir, id.String()))
}

func (m *JSONDir) CancelUpdate(id bakapy.TaskId) {
	m.unLockId(id)
}
