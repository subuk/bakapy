package bakapy

import (
	"fmt"
	"io/ioutil"
)

func NewTestMetaMan() MetaManager {
	tmpdir, err := ioutil.TempDir("", "metamantest_")
	if err != nil {
		panic(fmt.Errorf("cannot create temporary dir for test metaman:", err))
	}
	return NewMetaMan(&Config{MetadataDir: tmpdir})
}
