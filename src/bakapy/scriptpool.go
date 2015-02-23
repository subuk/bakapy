package bakapy

import (
	"fmt"
	"io/ioutil"
	"os"
	ospath "path"
)

type DirectoryScriptPool struct {
	root string
	tmp  string
}

func NewDirectoryScriptPool(config *Config) *DirectoryScriptPool {
	return &DirectoryScriptPool{root: config.CommandDir}
}

func (p *DirectoryScriptPool) BackupScript(name string) ([]byte, error) {
	scriptPath := ospath.Join(p.root, name)
	return ioutil.ReadFile(scriptPath)
}

func (p *DirectoryScriptPool) NotifyScript(name string) ([]byte, error) {
	scriptPath := ospath.Join(p.root, "notify-"+name+".sh")
	return ioutil.ReadFile(scriptPath)
}

func (p *DirectoryScriptPool) NotifyScriptPath(name string) (string, error) {
	content, err := p.NotifyScript(name)
	if err != nil {
		return "", err
	}
	newScriptFile, err := ioutil.TempFile(p.tmp, "notify-script"+name)
	if err != nil {
		return "", fmt.Errorf("cannot create temp file for notify script %s: %s", name, err)
	}
	defer newScriptFile.Close()

	_, err = newScriptFile.Write(content)
	if err != nil {
		return "", fmt.Errorf("cannot write notify script %s to temp file %s: %s", name, newScriptFile.Name(), err)
	}

	if err := os.Chmod(newScriptFile.Name(), 0700); err != nil {
		return "", fmt.Errorf("cannot change file permissions to 0700: %s", err)
	}
	return newScriptFile.Name(), nil
}
