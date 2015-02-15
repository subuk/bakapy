package bakapy

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type DummyFileInfo struct {
	name string
	size int64
}

func (i *DummyFileInfo) Name() string       { return i.name }     // base name of the file
func (i *DummyFileInfo) Size() int64        { return i.size }     // length in bytes for regular files; system-dependent for others
func (i *DummyFileInfo) Mode() os.FileMode  { return 0755 }       // file mode bits
func (i *DummyFileInfo) ModTime() time.Time { return time.Now() } // modification time
func (i *DummyFileInfo) IsDir() bool        { return false }      // abbreviation for Mode().IsDir()
func (i *DummyFileInfo) Sys() interface{}   { return nil }        // underlying data source (can return nil)

func TestStorage_CleanupExpired_Behavior(t *testing.T) {
	//
	// Test Storage.CleanupExpired behavior
	//
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	threeDays, _ := time.ParseDuration("72h")

	//
	// We have a storage
	//
	config := NewConfig()
	config.MetadataDir, _ = ioutil.TempDir("", "")
	config.StorageDir, _ = ioutil.TempDir("", "")
	defer os.RemoveAll(config.MetadataDir)
	defer os.RemoveAll(config.StorageDir)
	metaman := NewMetaMan(config)
	storage := NewStorage(config, metaman)
	os.MkdirAll(config.StorageDir+"/some_empty_dir", 0755)
	os.MkdirAll(config.StorageDir+"/some_restricted_dir", 0000)

	//
	// Active backup
	//
	os.MkdirAll(config.StorageDir+"/wow", 0755)
	f3, _ := os.Create(config.StorageDir + "/wow/file3.txt")
	f3.Close()
	f4, _ := os.Create(config.StorageDir + "/wow/file4.txt")
	f4.Close()
	metaman.Add("one", Metadata{
		Namespace:  "wow",
		ExpireTime: time.Now().Add(threeDays),
		Files: []MetadataFileEntry{
			{"file3.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
			{"file4.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
		},
	})

	//
	// Expired backup
	//
	os.MkdirAll(config.StorageDir+"/hello", 0755)
	f1, _ := os.Create(config.StorageDir + "/hello/file1.txt")
	f1.Close()
	f2, _ := os.Create(config.StorageDir + "/hello/file2.txt")
	f2.Close()
	metaman.Add("expired", Metadata{
		Namespace:  "hello",
		ExpireTime: time.Date(1970, 1, 1, 1, 1, 1, 1, time.UTC),
		Files: []MetadataFileEntry{
			{"file1.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
			{"file2.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
		},
	})

	//
	// Expired metadata with files deleted manually
	//
	metaman.Add("expired-broken", Metadata{
		Namespace:  "xxx",
		ExpireTime: time.Date(1970, 1, 1, 1, 1, 1, 1, time.UTC),
		Files: []MetadataFileEntry{
			{"file5.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
			{"file6.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
		},
	})

	//
	// And one corrupt metadata file
	//
	mdf_corrupt, _ := ioutil.TempFile(config.MetadataDir, "")
	mdf_corrupt.Write([]byte("{,wow'';'''"))

	//
	// After storage cleanup we does not expect any errors
	//
	err := CleanupExpiredJobs(metaman, storage)
	if err != nil {
		t.Fatal("error:", err)
	}

	//
	// Expect corrupt metadata file still exist
	//
	_, err = os.Stat(mdf_corrupt.Name())
	if err != nil {
		t.Fatal("cannot stat corrupt metadata file:", err)
	}

	//
	// Expect expired metadata file with manually deleted files still present
	//

	_, err = metaman.View("expired-broken")
	if err != nil {
		t.Fatal("cannot stat expired metadata file without data files:", err)
	}

	//
	// Expect expired backup removed (with metadata and files)
	//
	_, err = metaman.View("expired")
	if err == nil {
		t.Fatal("expired metadata file still present")
	}
	_, err = os.Stat(f1.Name())
	if err == nil {
		t.Fatal("data files for expired backup still present")
	}
	_, err = os.Stat(f2.Name())
	if err == nil {
		t.Fatal("data files for expired backup still present")
	}

	//
	// Expect active backup still present
	//
	_, err = metaman.View("one")
	if err != nil {
		t.Fatal("cannot stat active backup metadata file", err)
	}
	_, err = os.Stat(f3.Name())
	if err != nil {
		t.Fatal("cannot stat active backup data file", err)
	}
	_, err = os.Stat(f4.Name())
	if err != nil {
		t.Fatal("cannot stat active backup data file", err)
	}

}
