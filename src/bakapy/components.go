package bakapy

import (
	"io"
)

type Executer interface {
	Execute(script []byte, output io.Writer, errput io.Writer) error
}

type MetaManager interface {
	Keys() chan TaskId
	View(id TaskId) (Metadata, error)
	Add(id TaskId, md Metadata) error
	Update(id TaskId, up func(m *Metadata)) error
	AddFile(id TaskId, fm MetadataFileEntry) error
	Remove(id TaskId) error
}

type Notificator interface {
	JobFinished(md Metadata) error
	MetadataAccessFailed(err error) error
	Name() string
}

type BackupScriptPool interface {
	BackupScript(name string) ([]byte, error)
}

type NotifyScriptPool interface {
	NotifyScript(name string) ([]byte, error)
	NotifyScriptPath(name string) (string, error)
}

type ScriptPool interface {
	BackupScriptPool
	NotifyScriptPool
}
