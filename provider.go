package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Env struct{}

func (e *Env) Priority() int {
	return 30
}

func (e *Env) Provide(config interface{}) error {
	// TODO
	return nil
}

type Yaml struct{}

func (y *Yaml) Priority() int {
	return 50
}

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