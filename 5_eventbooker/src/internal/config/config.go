package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config is the configuration for the event booking service.
type Config struct {
	Srv       Server  `yaml:"server" env-required:"true"`
	Storage   Storage `yaml:"storage" env-required:"true"`
	Kafka     Kafka   `yaml:"kafka" env-required:"true"`
	NotifAddr string  `yaml:"notif_addr" env-required:"true"` // address of the delayed notification service
}

// Server is the configuration for the server.
type Server struct {
	Addrs []string `yaml:"addrs" env-required:"true"`
}

// Storage is the configuration for the storage.
type Storage struct {
	MasterDSN string   `yaml:"master_dsn" env-required:"true"`
	SlaveDSNs []string `yaml:"slave_dsns" env-required:"true"`
}

// Kafka is the configuration for the kafka.
type Kafka struct {
	Brokers []string `yaml:"brokers" env-required:"true"`
	Topic   string   `yaml:"topic" env-required:"true"`
	GroupID string   `yaml:"group_id" env-default:"eventbooker"`
}

// MustLoadConfig loads the config from the given path.
// if the path is empty, it will load the config from the default path "config.yaml".
// if the config is not valid, it will panic.
func MustLoadConfig(path string) *Config {
	if path == "" {
		path = "config.yaml"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file %s does not exist", path))
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
