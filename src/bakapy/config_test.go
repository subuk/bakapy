package bakapy

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
	"time"
)

var TEST_CONFIG_SYNTAX_ERROR = []byte(`
wow'''asdf
`)

var TEST_CONFIG_MAXAGE_BOTH = []byte(`
storage_dir: /tmp/backups/storage
metadata_dir: /tmp/backups/metadata
listen: 127.0.0.1:9876
jobs:
    wow:
      namespace: one
      max_age_days: 15
      max_age: 20m
`)

var TEST_CONFIG_MAXAGE_DAYS = []byte(`
storage_dir: /tmp/backups/storage
metadata_dir: /tmp/backups/metadata
listen: 127.0.0.1:9876
jobs:
    wow:
      namespace: one
      max_age_days: 30
`)

var TEST_CONFIG_MAXAGE = []byte(`
storage_dir: /tmp/backups/storage
metadata_dir: /tmp/backups/metadata
listen: 127.0.0.1:9876
jobs:
    wow:
      namespace: one
      max_age: 24h
`)

var TEST_CONFIG_WITHJOBS = []byte(`
storage_dir: /tmp/backups/storage
metadata_dir: /tmp/backups/metadata
listen: 127.0.0.1:9876
jobs:
    wow:
      namespace: one
`)

var JOBS_CONFIG = []byte(`
xxx:
  namespace: one
zzz:
  namespace: two
`)

var JOBS_CONFIG_ONE = []byte(`
yyy:
  namespace: one
`)

var TEST_CONFIG_REMOTE_FILTERS = []byte(`
storage_dir: /tmp/backups/storage
metadata_dir: /tmp/backups/metadata
listen: 127.0.0.1:9876
args:
    vhosts_dir: /home/vhosts
    backup_type: full
    listed_incremental_dir: /tmp/backup-meta
jobs:
    xxx:
      namespace: one
      remote_filters:
        - gzip: {level: 1}
        - encrypt: {key: "secewr21!@$98"}
`)

var TEST_CONFIG_JOB_WITH_REMOTE_FILTERS = []byte(`
storage_dir: /tmp/backups/storage
metadata_dir: /tmp/backups/metadata
command_dir: /etc/bakapy/commands
listen: 127.0.0.1:9876
args:
    vhosts_dir: /home/vhosts
    backup_type: full
    listed_incremental_dir: /tmp/backup-meta
jobs:
    xxx:
      namespace: one
      host: user@192.168.10.149
      remote_filters:
        - gzip: {level: 1, force: true}
        - encrypt: {key: "secewr21!@$98"}
        - enlarge: {}
`)

var TEST_CONFIG_TEMPLATE_DIRECTORY = []byte(`
storage_dir: /tmp/backups/storage
metadata_dir: /tmp/backups/metadata
jobs:
    zzz:
      namespace: two
      temp_dir: ~/.bakapy_filters
`)

func TestParseConfig_WithJobs(t *testing.T) {
	cfg, _ := ioutil.TempFile("", "test_config")
	cfg.Write(TEST_CONFIG_WITHJOBS)
	cfg.Close()
	defer os.Remove(cfg.Name())

	config, err := ParseConfig(cfg.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(config.Jobs) != 1 {
		t.Fatal("config.Jobs length != 1 | ", len(config.Jobs))
	}

	jConfig := config.Jobs["wow"]
	if jConfig == nil {
		t.Fatal("jConfig must not be nil")
	}
}

func TestParseConfig_IncludeJobsRelative(t *testing.T) {

	jobsConfig, _ := ioutil.TempFile("", "jobsconfig")
	jobsConfig.Write(JOBS_CONFIG)
	jobsConfig.Close()
	defer os.Remove(jobsConfig.Name())

	mainConfig, _ := ioutil.TempFile("", "testconfig")
	mainConfig.Write([]byte("include_jobs: [jobsconfig*]"))
	mainConfig.Close()
	defer os.Remove(mainConfig.Name())

	config, err := ParseConfig(mainConfig.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(config.Jobs) != 2 {
		t.Fatal("config.Jobs length != 2 | ", len(config.Jobs))
	}
}

func TestParseConfig_IncludeJobsDuplicates(t *testing.T) {

	jobsConfig, _ := ioutil.TempFile("", "jobsconfig1")
	jobsConfig.Write(JOBS_CONFIG_ONE)
	jobsConfig.Close()
	defer os.Remove(jobsConfig.Name())

	jobsConfig2, _ := ioutil.TempFile("", "jobsconfig2")
	jobsConfig2.Write(JOBS_CONFIG_ONE)
	jobsConfig2.Close()
	defer os.Remove(jobsConfig2.Name())

	mainConfig, _ := ioutil.TempFile("", "testconfig")
	mainConfig.Write([]byte("include_jobs: [jobsconfig1*,jobsconfig2*]"))
	mainConfig.Close()
	defer os.Remove(mainConfig.Name())

	_, err := ParseConfig(mainConfig.Name())
	expectedErr := fmt.Sprintf("%s: duplicated job name %s, previously defined at %s", jobsConfig2.Name(), "yyy", jobsConfig.Name())
	if err.Error() != expectedErr {
		t.Fatal(err, "|", expectedErr)
	}
}

func TestParseConfig_MaxAgeDaysSanitizationOk(t *testing.T) {

	cfg, _ := ioutil.TempFile("", "test_config")
	cfg.Write(TEST_CONFIG_MAXAGE_DAYS)
	cfg.Close()
	defer os.Remove(cfg.Name())

	config, err := ParseConfig(cfg.Name())
	if err != nil {
		t.Fatal(err)
	}

	jConfig := config.Jobs["wow"]
	if jConfig == nil {
		t.Fatal("jConfig must not be nil")
	}

	if jConfig.MaxAge != time.Duration(time.Hour*24*30) {
		t.Fatal("Wrong jConfig.MaxAge. Must be", time.Duration(time.Hour*24*30))
	}
}

func TestParseConfig_MaxAgeDaysSanitizationFailBothSpecified(t *testing.T) {

	cfg, _ := ioutil.TempFile("", "test_config")
	cfg.Write(TEST_CONFIG_MAXAGE_BOTH)
	cfg.Close()
	defer os.Remove(cfg.Name())

	_, err := ParseConfig(cfg.Name())
	expectedErr := "job wow: both max_age and max_age_days defined. max_age='20m0s' max_age_days='15'"
	if err.Error() != expectedErr {
		t.Fatal(err, "| != |", expectedErr)
	}
}

func TestParseConfig_FileDoesNotExist(t *testing.T) {
	_, err := ParseConfig("DOES_NOT_EXIST")
	expectedErr := "open DOES_NOT_EXIST: no such file or directory"
	if err.Error() != expectedErr {
		t.Fatal(err, "| != |", expectedErr)
	}
}

func TestParseConfig_SyntaxError(t *testing.T) {
	cfg, _ := ioutil.TempFile("", "test_config")
	cfg.Write(TEST_CONFIG_SYNTAX_ERROR)
	cfg.Close()
	defer os.Remove(cfg.Name())

	_, err := ParseConfig(cfg.Name())
	expectedErr := "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!str `wow'''asdf` into bakapy.Config"
	if err.Error() != expectedErr {
		t.Fatal(err, "| != |", expectedErr)
	}
}

func TestParseConfig_MaxAgeDaysParseOk(t *testing.T) {

	cfg, _ := ioutil.TempFile("", "test_config")
	cfg.Write(TEST_CONFIG_MAXAGE)
	cfg.Close()
	defer os.Remove(cfg.Name())

	config, err := ParseConfig(cfg.Name())
	if err != nil {
		t.Fatal(err)
	}

	jConfig := config.Jobs["wow"]
	if jConfig == nil {
		t.Fatal("jConfig must not be nil")
	}

	if jConfig.MaxAge != time.Duration(time.Hour*24) {
		t.Fatal("Wrong jConfig.MaxAge.", jConfig.MaxAge, "Must be", time.Duration(time.Hour*24))
	}
}

func TestRunAtSpec_SchedulerString_NoSecond(t *testing.T) {
	spec := &RunAtSpec{
		Minute:  "3",
		Hour:    "44",
		Day:     "*",
		Month:   "*",
		Weekday: "*",
	}
	s := spec.SchedulerString()
	if s != "0 3 44 * * *" {
		t.Fatal("Must be '0 3 44 * * *' not ", s)
	}
}

func TestRunAtSpec_SchedulerString_WithSecond(t *testing.T) {
	spec := &RunAtSpec{
		Second:  "4",
		Minute:  "3",
		Hour:    "44",
		Day:     "*",
		Month:   "*",
		Weekday: "*",
	}
	s := spec.SchedulerString()
	if s != "4 3 44 * * *" {
		t.Fatal("Must be '4 3 44 * * *' not ", s)
	}
}

func TestParseConfig_JobsWithRemoteFilters(t *testing.T) {
	cfg, _ := ioutil.TempFile("", "test_config")
	cfg.Write(TEST_CONFIG_REMOTE_FILTERS)
	cfg.Close()
	defer os.Remove(cfg.Name())

	config, err := ParseConfig(cfg.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(config.Jobs["xxx"].RemoteFilters) != 2 {
		t.Fatal("RemoteFilters length != 2 | ", len(config.Jobs["xxx"].RemoteFilters))
	}

	if _, exist := config.Jobs["xxx"].RemoteFilters[0]["gzip"]; !exist {
		t.Fatal("Remote filters does not parsed")
	}
	if value := config.Jobs["xxx"].RemoteFilters[0]["gzip"].Params["level"]; value != "1" {
		t.Fatalf(`Remote filters does not parse values properly: "%s"; want "1"`)
	}
}

func TestParseConfig_TemplateDirectory(t *testing.T) {
	cfg, _ := ioutil.TempFile("", "test_config")
	cfg.Write(TEST_CONFIG_TEMPLATE_DIRECTORY)
	cfg.Close()
	defer os.Remove(cfg.Name())

	config, err := ParseConfig(cfg.Name())
	if err != nil {
		t.Fatal(err)
	}

	if config.Jobs["zzz"].TempDir != "~/.bakapy_filters" {
		t.Fatalf(`Wrong template directory: "%s"; expected "~/.bakapy_filters"`, config.Jobs["zzz"].TempDir)
	}
}

func TestMakeRemoteFiltersOnClient(t *testing.T) {
	cfg, _ := ioutil.TempFile("", "test_config")
	cfg.Write(TEST_CONFIG_JOB_WITH_REMOTE_FILTERS)
	cfg.Close()
	defer os.Remove(cfg.Name())

	config, err := ParseConfig(cfg.Name())
	if err != nil {
		t.Fatal(err)
	}

	expectedFiltersCmds := []*regexp.Regexp{
		regexp.MustCompile(`[ARG_FORCE=true|ARG_LEVEL=1]{2} /tmp/filter[a-zA-Z0-9]{12}\.sh`),
		regexp.MustCompile(`ARG_KEY=secewr21!@\$98 /tmp/filter[a-zA-Z0-9]{12}\.sh`),
		regexp.MustCompile(`/tmp/filter[a-zA-Z0-9]{12}\.sh`),
	}

	filters, err := config.Jobs["xxx"].RemoteFilters.MakeFiltersOnClient(config.Jobs["xxx"].Host, config.CommandDir, false)
	if err != nil {
		t.Fatalf("%s\n", err)
	}

	if len(filters) != 3 {
		t.Fatalf("Unexpected filters quantity: %d\n", len(filters))
	}

	for i := 0; i < len(filters); i++ {
		if !expectedFiltersCmds[i].MatchString(filters[i]) {
			t.Fatalf("Unexpected filter command: %s\n", filters[i])
		}
	}
}
