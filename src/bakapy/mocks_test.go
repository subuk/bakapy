package bakapy

import (
	"fmt"
	"io/ioutil"
	"os"
)

func NewTestMetaMan() MetaManager {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(fmt.Errorf("cannot create temporary dir for test metaman:", err))
	}
	defer os.RemoveAll(tmpdir)
	return NewMetaMan(&Config{MetadataDir: tmpdir})
}
