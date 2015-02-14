package bakapy

import (
	"fmt"
	"io/ioutil"
	"os"
)

func NewTestMetaMan() MetaManager {
	tmpdir, err := ioutil.TempDir("", "metamantest_")
	if err != nil {
		panic(fmt.Errorf("cannot create temporary dir for test metaman:", err))
	}
	return NewMetaMan(&Config{MetadataDir: tmpdir})
}

type TestScriptPool struct {
	Err      error
	Script   []byte
	TempFile string
}

func (t *TestScriptPool) BackupScript(name string) ([]byte, error) {
	return t.Script, t.Err
}

func (t *TestScriptPool) NotifyScript(name string) ([]byte, error) {
	return t.Script, t.Err
}

func (t *TestScriptPool) NotifyScriptPath(name string) (string, error) {
	if t.TempFile == "" {
		return "", fmt.Errorf("TestScriptPool.TempFile must be set")
	}
	content, err := t.NotifyScript(name)
	if err != nil {
		return "", err
	}
	os.Chmod(t.TempFile, 0755)
	ioutil.WriteFile(t.TempFile, content, 0755)
	return t.TempFile, nil
}
