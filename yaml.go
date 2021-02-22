package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Yaml is a provider for configuration using yaml file
type Yaml struct{}

// Priority of yaml provider, set to 50
func (y *Yaml) Priority() int {
	return 50
}

// Provide loads configuration from yaml file
func (y *Yaml) Provide(config interface{}) error {
	b, err := readFile()
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, config)
	if err != nil {
		return err
	}

	return nil
}

func readFile() ([]byte, error) {
	dir, err := execDir()
	p, err := filepath.Abs(dir + "/config.yaml")
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func execDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(ex), nil
}
