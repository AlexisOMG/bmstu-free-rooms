package main

import (
	"fmt"
	"ics/database"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Database    *database.Config `yaml:"database"`
	ScheduleDir *string          `yaml:"schedule_dir"`
}

func readConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	config := &Config{}
	err = yaml.NewDecoder(file).Decode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	return config, nil
}
