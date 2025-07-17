package model

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

const (
	DefaultWritePermissions = 0700
	ConfigFileName          = "config.yml"
)

var CliInstanceId string

type FablabConfig struct {
	Instances  map[string]*InstanceConfig `yaml:"instances"`
	Default    string                     `yaml:"default"`
	ConfigPath string                     `yaml:"-"`
}

func (self *FablabConfig) GetSelectedInstanceId() string {
	if CliInstanceId != "" {
		return CliInstanceId
	}
	if self.Default != "" {
		return self.Default
	}
	return "default"
}

type InstanceConfig struct {
	Id               string `yaml:"name"`
	Model            string `yaml:"model"`
	WorkingDirectory string `yaml:"working_directory"`
	Executable       string `yaml:"executable"`
}

func (self *InstanceConfig) CleanupWorkingDir() error {
	if err := os.RemoveAll(self.WorkingDirectory); err != nil {
		return errors.Wrapf(err, "error cleaning up instance [%s] with working directory %v", self.Id, self.WorkingDirectory)
	}
	return nil
}

func (self *InstanceConfig) LoadLabel() (*Label, error) {
	return LoadLabel(self.WorkingDirectory)
}

func GetConfig() *FablabConfig {
	if config == nil {
		var err error
		config, err = loadConfig()
		if err != nil {
			if config == nil {
				logrus.WithError(err).Fatalf("unable to load configuration")
			} else {
				logrus.WithError(err).Fatalf("unable to load configuration at %v", config.ConfigPath)
			}
		}
	}
	return config
}

func loadConfig() (*FablabConfig, error) {
	cfgDir, err := ConfigDir()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get config dir while loading cli configuration")
	}
	configFile := filepath.Join(cfgDir, ConfigFileName)
	return LoadConfig(configFile)
}

func LoadConfig(configFile string) (*FablabConfig, error) {
	config := &FablabConfig{
		Instances: map[string]*InstanceConfig{},
	}

	config.ConfigPath = configFile

	_, err := os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, errors.Wrapf(err, "error while statting config file %v", configFile)
	}
	result, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading config file %v", configFile)
	}

	if err := yaml.Unmarshal(result, config); err != nil {
		return nil, errors.Wrapf(err, "error while parsing YAML config file %v", configFile)
	}

	if config.Instances == nil {
		config.Instances = map[string]*InstanceConfig{}
	}
	return config, nil
}

func PersistConfig(config *FablabConfig) error {
	if config.Default == "" {
		config.Default = "default"
	}

	if config.Instances[config.Default] == nil {
		for k := range config.Instances {
			config.Default = k
			break
		}
	}

	cfgDir, err := ConfigDir()
	if err != nil {
		return errors.Wrap(err, "couldn't get config dir while persisting fablab configuration")
	}
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		return errors.Wrapf(err, "unable to create config dir %v", cfgDir)
	}

	configFile := filepath.Join(cfgDir, ConfigFileName)

	data, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "error while marshalling config to YAML")
	}

	err = os.WriteFile(configFile, data, 0600)
	if err != nil {
		return errors.Wrapf(err, "error while writing config file %v", configFile)
	}

	return nil
}

func GetActiveInstanceConfig() *InstanceConfig {
	_, err := loadActiveInstanceConfig()
	if err != nil {
		logrus.WithError(err).Fatal("error loading active instance config")
	}
	return instanceConfig
}

func loadActiveInstanceConfig() (*InstanceConfig, error) {
	if instanceConfig == nil {
		GetConfig()
		id := config.GetSelectedInstanceId()
		instance, found := config.Instances[id]
		if !found {
			return nil, errors.Errorf("no identity '%v' found in cli config %v", id, config.ConfigPath)
		}
		instanceConfig = instance
	}
	return instanceConfig, nil
}

func ConfigDir() (string, error) {
	path := os.Getenv("FABLAB_HOME")
	if path != "" {
		return path, nil
	}

	h := HomeDir()
	path = filepath.Join(h, ".fablab")

	err := os.MkdirAll(path, DefaultWritePermissions)
	if err != nil {
		return "", err
	}
	return path, nil
}

func HomeDir() string {
	h, err := os.UserHomeDir()
	if err == nil {
		return h
	}
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	h = os.Getenv("USERPROFILE") // windows
	if h == "" {
		h = "."
	}
	return h
}
