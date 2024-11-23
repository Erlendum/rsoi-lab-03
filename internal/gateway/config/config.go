package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Server struct {
	Address         string        `yaml:"address"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type Config struct {
	Server               Server `yaml:"server"`
	ReservationSystemURL string `yaml:"reservation_system_url"`
	LibrarySystemURL     string `yaml:"library_system_url"`
	RatingSystemURL      string `yaml:"rating_system_url"`
}

func New() (*Config, error) {
	cfg := &Config{}

	yamlFile, err := os.ReadFile(fmt.Sprint("./configs/gateway/config.yml"))
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}
