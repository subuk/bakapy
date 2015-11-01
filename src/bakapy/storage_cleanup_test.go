package bakapy

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

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
	storage := NewStorage(config)
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
	m2f, _ := ioutil.TempFile(config.MetadataDir, "")
	m2f.Close()
	(&JobMetadata{
		TaskId:     "one",
		Namespace:  "wow",
		JobName:    "testjob1",
		Success:    true,
		ExpireTime: time.Now().Add(threeDays),
		Files: []JobMetadataFile{
			{"file3.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
			{"file4.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
		},
	}).Save(m2f.Name())

	//
	// Expired backup
	//
	os.MkdirAll(config.StorageDir+"/hello", 0755)
	f1, _ := os.Create(config.StorageDir + "/hello/file1.txt")
	f1.Close()
	f2, _ := os.Create(config.StorageDir + "/hello/file2.txt")
	f2.Close()
	m1f, _ := ioutil.TempFile(config.MetadataDir, "")
	m1f.Close()
	(&JobMetadata{
		Namespace:  "hello",
		JobName:    "testjob2",
		Success:    true,
		ExpireTime: time.Date(1970, 1, 1, 1, 1, 1, 1, time.UTC),
		Files: []JobMetadataFile{
			{"file1.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
			{"file2.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
		},
	}).Save(m1f.Name())

	//
	// Expired metadata with files deleted manually
	//
	m3f, _ := ioutil.TempFile(config.MetadataDir, "")
	m3f.Close()
	(&JobMetadata{
		Namespace:  "xxx",
		JobName:    "testjob2",
		Success:    true,
		ExpireTime: time.Date(1970, 1, 1, 1, 1, 1, 1, time.UTC),
		Files: []JobMetadataFile{
			{"file5.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
			{"file6.txt", 0, "1.1.1.1", (time.Time{}), (time.Time{})},
		},
	}).Save(m3f.Name())

	//
	// And one corrupt metadata file
	//
	m4f_corrupt, _ := ioutil.TempFile(config.MetadataDir, "")
	m4f_corrupt.Write([]byte("{,wow'';'''"))

	//
	// After storage cleanup we does not expect any errors
	//
	err := storage.CleanupExpired()
	if err != nil {
		t.Fatal("error:", err)
	}

	//
	// Expect corrupt metadata file moved to other directory
	//
	_, m4f_corrupt_filename := path.Split(m4f_corrupt.Name())
	_, err = os.Stat(path.Join(config.MetadataDir+"_corrupted", m4f_corrupt_filename))
	if err != nil {
		t.Fatal("cannot stat corrupt metadata file:", err)
	}

	//
	// Expect expired metadata file with manually deleted files removed
	//
	_, err = os.Stat(m3f.Name())
	if err == nil {
		t.Fatal("expired metadata file without data files still present", err)
	}

	//
	// Expect expired backup removed (with metadata and files)
	//
	_, err = os.Stat(m1f.Name())
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
	_, err = os.Stat(m2f.Name())
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
