package main

import (
	"eventbooker/internal/config"
	"os"

	"github.com/kxddry/wbf/zlog"
)

func main() {
	var cfg config.Config
	zlog.Init()
	log := zlog.Logger
	if err := cfg.Update("config.yaml"); err != nil {
		log.Warn().Err(err).Msg("failed to load config")
	}
	if err := cfg.Update(os.Getenv("CONFIG_FILE")); err != nil {
		log.Warn().Err(err).Msg("failed to load config")
	}

}
