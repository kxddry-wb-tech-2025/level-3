package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config is the configuration for the salestracker service.
type Config struct {
	Storage StorageConfig `yaml:"storage"`
	Server  ServerConfig  `yaml:"server"`
}

// StorageConfig is the configuration for the storage.
type StorageConfig struct {
	Host     string `yaml:"host" env:"STORAGE_HOST" env-required:"true"`
	Port     string `yaml:"port" env:"STORAGE_PORT" env-required:"true"`
	Username string `yaml:"username" env:"STORAGE_USERNAME" env-required:"true"`
	Password string `yaml:"password" env:"STORAGE_PASSWORD" env-required:"true"`
	Database string `yaml:"database" env:"STORAGE_DATABASE" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env:"STORAGE_SSLMODE" env-required:"true"`
}

// ServerConfig is the configuration for the server.
type ServerConfig struct {
	Addrs     []string `yaml:"addrs" env-required:"false"`
	StaticDir string   `yaml:"static_dir" env-required:"false"`
}

// New creates a new Config instance.
func New(path string) (*Config, error) {
	if path == "" {
		path = "config.yaml"
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// GetMasterDSN returns the master DSN for the storage.
func (c *Config) GetMasterDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.Storage.Username, c.Storage.Password, c.Storage.Host, c.Storage.Port, c.Storage.Database, c.Storage.SSLMode)
}
