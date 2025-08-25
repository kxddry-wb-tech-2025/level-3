package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Storage StorageConfig `yaml:"storage"`
}

type StorageConfig struct {
	Host     string `yaml:"host" env:"STORAGE_HOST" env-required:"true"`
	Port     string `yaml:"port" env:"STORAGE_PORT" env-required:"true"`
	Username string `yaml:"username" env:"STORAGE_USERNAME" env-required:"true"`
	Password string `yaml:"password" env:"STORAGE_PASSWORD" env-required:"true"`
	Database string `yaml:"database" env:"STORAGE_DATABASE" env-required:"true"`
}

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
