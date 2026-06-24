package main

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type rootConfig struct {
	Parser string `yaml:"parser"`
}

func readParserName(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var cfg rootConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return "", err
	}

	if cfg.Parser == "" {
		return "", errors.New("config missing required field: parser")
	}

	return cfg.Parser, nil
}