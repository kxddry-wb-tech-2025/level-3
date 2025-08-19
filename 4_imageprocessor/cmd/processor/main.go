package main

import (
	"os"

	"github.com/kxddry/wbf/config"
	"github.com/kxddry/wbf/zlog"
)

func main() {
	zlog.Init()
	cfg := config.New()

	if err := cfg.Load(os.Getenv("CONFIG_PATH")); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to load config with CONFIG_PATH")
	}
	if err := cfg.Load("config.yaml"); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to load config with config.yaml")
	}
}
