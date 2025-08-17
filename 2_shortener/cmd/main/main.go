package main

import (
	"context"
	"shortener/internal/storage/postgres"

	"github.com/kxddry/wbf/config"
	"github.com/kxddry/wbf/zlog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zlog.Init()

	cfg := config.New()
	cfg.Load("./config.yaml")

	storage, err := postgres.New(cfg.GetString("postgres.master"), nil)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer storage.Close()
}
