package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	cfg "github.com/kxddry/wbf/config"
)

// Config is the configuration for the event booking service.
type Config struct {
	Srv         Server
	BookTimeout time.Duration
	CronTicker  time.Duration
	Storage     Storage
}

// Server is the configuration for the server.
type Server struct {
	Addr string
}

// Storage is the configuration for the storage.
type Storage struct {
	MasterDSN string
	SlaveDSNs []string
}

// Update updates the configuration from a file.
func (c *Config) Update(file string) error {
	if c == nil {
		c = &Config{}
	}
	cc := cfg.New()
	err := cc.Load(file)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %w", err)
		}
		return err
	}

	tmpAddr := cc.GetString("server.addr")
	if tmpAddr != "" {
		c.Srv.Addr = tmpAddr
	}

	tmpBookTimeoutStr := cc.GetString("book.timeout")
	if tmpBookTimeoutStr != "" {
		tmpBookTimeout, err := time.ParseDuration(tmpBookTimeoutStr)
		if err == nil {
			c.BookTimeout = tmpBookTimeout
		}
	}

	tmpCronTickerStr := cc.GetString("cron.ticker")
	if tmpCronTickerStr != "" {
		tmpCronTicker, err := time.ParseDuration(tmpCronTickerStr)
		if err == nil {
			c.CronTicker = tmpCronTicker
		}
	}

	tmpStorageMasterDSN := cc.GetString("storage.master_dsn")
	if tmpStorageMasterDSN != "" {
		c.Storage.MasterDSN = tmpStorageMasterDSN
	}

	tmpStorageSlaveDSNs := cc.GetString("storage.slave_dsns")
	if tmpStorageSlaveDSNs != "" {
		c.Storage.SlaveDSNs = strings.Split(tmpStorageSlaveDSNs, ",")
	}

	return nil
}

func New() *Config {
	return &Config{
		Srv: Server{
			Addr: ":8080",
		},
		BookTimeout: 20 * time.Minute,
		CronTicker:  15 * time.Minute,
		Storage: Storage{
			MasterDSN: os.ExpandEnv("postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}"),
			SlaveDSNs: nil,
		},
	}
}
