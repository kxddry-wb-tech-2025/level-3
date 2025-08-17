package main

import (
	"context"
	"os"
	"shortener/internal/api"
	"shortener/internal/storage/postgres"
	"shortener/internal/validator"

	"github.com/kxddry/wbf/config"
	"github.com/kxddry/wbf/zlog"
	"github.com/subosito/gotenv"
)

func main() {
	_ = gotenv.Load(".env")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zlog.Init()

	cfg := config.New()
	if err := cfg.Load(os.Getenv("CONFIG_PATH")); err != nil {
		zlog.Logger.Info().Msg("CONFIG_PATH is not set, searching for config.yaml in the current directory")
		if err := cfg.Load("./config.yaml"); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to load config")
			os.Exit(1)
		}
	}

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
