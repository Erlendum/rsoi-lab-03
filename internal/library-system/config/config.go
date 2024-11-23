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

type PostgreSQL struct {
	DSN string `env:"POSTGRESQL_DSN"`
}

type Config struct {
	Server     Server `yaml:"server"`
	PostgreSQL PostgreSQL
}

func New() (*Config, error) {
	cfg := &Config{}

	cfg.PostgreSQL.DSN = os.Getenv("POSTGRESQL_DSN")

	yamlFile, err := os.ReadFile(fmt.Sprint("./configs/library-system/config.yml"))
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}
