package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/AlexisOMG/bmstu-free-rooms/database"
)

type Config struct {
	Database    *database.Config `yaml:"database"`
	ScheduleDir *string          `yaml:"schedule_dir"`
	Token       *string          `yaml:"bot_token"`
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
