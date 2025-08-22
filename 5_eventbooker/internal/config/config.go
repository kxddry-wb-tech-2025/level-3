package config

import (
	"fmt"
	"os"
	"time"

	cfg "github.com/kxddry/wbf/config"
)

type Config struct {
	Srv         Server
	BookTimeout time.Duration
	CronTicker  time.Duration
	StorageDSN  string
}

type Server struct {
	Addr string
}

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

	tmpStorageDSN := os.ExpandEnv(cc.GetString("storage.dsn"))
	if tmpStorageDSN != "" {
		c.StorageDSN = tmpStorageDSN
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
		StorageDSN:  os.ExpandEnv("postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}"),
	}
}
