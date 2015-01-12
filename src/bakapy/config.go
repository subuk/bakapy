package bakapy

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"path/filepath"
	"time"
)

type Config struct {
	IncludeJobs []string `yaml:"include_jobs"`
	Listen      string
	StorageDir  string     `yaml:"storage_dir"`
	MetadataDir string     `yaml:"metadata_dir"`
	CommandDir  string     `yaml:"command_dir"`
	SMTP        SMTPConfig `yaml:"smtp"`
	Jobs        map[string]*JobConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	MailFrom string
	MailTo   string
}

func (cfg *Config) PrettyFmt() []byte {
	s, _ := yaml.Marshal(cfg)
	return s
}

type RunAtSpec struct {
	Second  string
	Minute  string
	Hour    string
	Day     string
	Month   string
	Weekday string
}

func (r *RunAtSpec) SchedulerString() string {
	if r.Second == "" {
		r.Second = "0"
	}
	return fmt.Sprintf(
		"%s %s %s %s %s %s",
		r.Second,
		r.Minute,
		r.Hour,
		r.Day,
		r.Month,
		r.Weekday,
	)
}

type JobConfig struct {
	Sudo       bool
	Disabled   bool
	Gzip       bool
	MaxAgeDays int           `yaml:"max_age_days"`
	MaxAge     time.Duration `yaml:"max_age"`
	Namespace  string
	Host       string
	Port       uint
	Command    string
	Args       map[string]string
	RunAt      RunAtSpec `yaml:"run_at"`
	executor   Executer  `yaml:"-"`
}

func (jobConfig *JobConfig) Sanitize() error {
	if jobConfig.MaxAgeDays != 0 && jobConfig.MaxAge != 0 {
		e := fmt.Sprintf("both max_age and max_age_days defined. max_age='%s' max_age_days='%d'",
			jobConfig.MaxAge, jobConfig.MaxAgeDays)
		return errors.New(e)
	}
	if jobConfig.MaxAgeDays != 0 {
		jobConfig.MaxAge = time.Duration(jobConfig.MaxAgeDays) * time.Hour * 24
	}
	return nil
}

func NewConfig() *Config {
	jobs := Config{
		Jobs: map[string]*JobConfig{},
	}
	return &jobs
}

func ParseConfig(configPath string) (*Config, error) {
	cfg := NewConfig()

	rawConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(rawConfig, cfg)
	if err != nil {
		return nil, err
	}

	configDir := path.Dir(configPath)
	jobDefines := map[string]string{}
	for _, relPathGlob := range cfg.IncludeJobs {
		pathGlob := path.Join(configDir, relPathGlob)
		paths, err := filepath.Glob(pathGlob)
		if err != nil {
			return nil, err
		}

		for _, path := range paths {
			raw, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, err
			}
			jobs := map[string]*JobConfig{}
			err = yaml.Unmarshal(raw, &jobs)
			if err != nil {
				return nil, err
			}
			for name, params := range jobs {
				if _, exist := jobDefines[name]; exist {
					errString := fmt.Sprintf(
						"%s: duplicated job name %s, previously defined at %s",
						path, name, jobDefines[name])
					return nil, errors.New(errString)
				}
				jobDefines[name] = path
				cfg.Jobs[name] = params
			}
		}
	}

	for jobName, jobConfig := range cfg.Jobs {
		err := jobConfig.Sanitize()
		if err != nil {
			return nil, errors.New("job " + jobName + ": " + err.Error())
		}
	}

	return cfg, nil
}
