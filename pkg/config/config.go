package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
)

const configFileName = ".k8swatch.yaml"

// Resource struct: resource configuration
type Resource struct {
	Pod bool `json:"po"`
	// will add other resources later
}

// Handler struct:
type Handler struct {
}

// Config struct: k8swatch configuration
type Config struct {
	Handler   Handler  `json:"handler"`
	Resource  Resource `json:"resource"`
	Namespace string   `json:"namespace"`
}

// New creates new config object
func New() (*Config, error) {
	c := &Config{}
	err := c.Load()
	return c, err
}

// create k8swatch config file if not exist
func createIfNotExist() error {
	configFile := filepath.Join(configDir(), configFileName)
	_, err := os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(configFile)
			if err != nil {
				return err
			}
			file.Close()
		} else {
			return err
		}
	}
	return nil
}

// Load loads configuration from config file
func (c *Config) Load() error {
	err := createIfNotExist()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir(), configFileName)
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if len(b) != 0 {
		return yaml.Unmarshal(b, c)
	}

	return nil

}

func configDir() string {
	if configDir := os.Getenv("KW_CONFIG"); configDir != "" {
		return configDir
	}

	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}

	return os.Getenv("HOME")
}
