package config

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Yaml is a provider for configuration using yaml file
type Yaml struct {
	Path string
}

// Provide loads configuration from yaml file
func (y *Yaml) Provide(config interface{}) error {
	b, err := y.readFile()
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, config)
	if err != nil {
		return err
	}

	return nil
}

func (y *Yaml) readFile() ([]byte, error) {
	p, err := y.resolvePath()
	if err != nil {
		return nil, err
	}

	return os.ReadFile(p)
}

func (y *Yaml) resolvePath() (string, error) {
	if filepath.IsAbs(y.Path) {
		return y.Path, nil
	}

	dir, err := execDir()
	if err != nil {
		return "", err
	}

	if y.Path != "" {
		return filepath.Join(dir, y.Path), nil
	}

	return filepath.Join(dir, "config.yaml"), nil
}

func execDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(ex), nil
}
