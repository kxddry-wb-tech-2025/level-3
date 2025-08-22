package main

import (
	"eventbooker/src/internal/config"
	"os"

	"github.com/kxddry/wbf/zlog"
)

func main() {
	cfg := config.New()
	zlog.Init()
	log := zlog.Logger
	if err := cfg.Update("config.yaml"); err != nil {
		log.Warn().Err(err).Msg("failed to load config")
	}
	if err := cfg.Update(os.Getenv("CONFIG_FILE")); err != nil {
		log.Warn().Err(err).Msg("failed to load config")
	}

}
