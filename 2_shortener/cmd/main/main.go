package main

import (
	"context"
	"shortener/internal/api"
	"shortener/internal/storage/postgres"
	"shortener/internal/validator"

	"github.com/kxddry/wbf/config"
	"github.com/kxddry/wbf/zlog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zlog.Init()

	cfg := config.New()
	cfg.Load("./config.yaml")

	store, err := postgres.New(ctx, cfg.GetString("postgres.master"), nil)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer store.Close()

	v := validator.New()
	srv := api.New(store, store, *v)
	// Register routes (with ctx for async click logging)
	srv.RegisterRoutes(ctx)

	if err := srv.Run(ctx); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("server exited with error")
	}
}
