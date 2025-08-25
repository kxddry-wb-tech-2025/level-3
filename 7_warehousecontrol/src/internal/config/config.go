package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config is the configuration for the warehousecontrol service.
type Config struct {
	Storage StorageConfig `yaml:"storage"`
	Server  ServerConfig  `yaml:"server"`
	JWT     JWTConfig     `yaml:"jwt"`
}

// JWTConfig is the configuration for the JWT.
type JWTConfig struct {
	Secret     string `yaml:"secret" env:"SECRET" env-required:"true"`
	CookieName string `yaml:"cookie_name" env:"COOKIE_NAME" env-default:"auth"`
}

// StorageConfig is the configuration for the storage.
type StorageConfig struct {
	MasterDSN string   `yaml:"master_dsn"`
	SlaveDSNs []string `yaml:"slave_dsns"`
}

// ServerConfig is the configuration for the server.
type ServerConfig struct {
	Address   string `yaml:"address"`
	StaticDir string `yaml:"static_dir"`
}

// New creates a new Config instance.
func MustLoad(path string) *Config {
	if path == "" {
		path = "./config/app.yaml"
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic(err)
	}

	for i := range cfg.Storage.SlaveDSNs {
		cfg.Storage.SlaveDSNs[i] = os.ExpandEnv(cfg.Storage.SlaveDSNs[i])
	}

	cfg.Storage.MasterDSN = os.ExpandEnv(cfg.Storage.MasterDSN)

	return &cfg
}
