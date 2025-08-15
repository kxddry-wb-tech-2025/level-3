package main

import (
	"os"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()
	log := zlog.Logger
	log.Debug().Msg("debug enabled")

	cfg := config.New()
	err := cfg.Load("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config.yaml")
		os.Exit(1)
	}

}
