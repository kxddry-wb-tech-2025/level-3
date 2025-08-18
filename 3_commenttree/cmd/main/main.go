package main

import (
	"comment-tree/internal/api"
	"comment-tree/internal/storage/postgres"
	"os"
	"os/signal"
	"syscall"

	"github.com/kxddry/wbf/config"
	"github.com/kxddry/wbf/dbpg"
	"github.com/kxddry/wbf/zlog"
	"github.com/subosito/gotenv"
)

func main() {
	gotenv.Load(".env")
	zlog.Init()
	cfg := config.New()
	if err := cfg.Load(os.Getenv("CONFIG_PATH")); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to load config with CONFIG_PATH")
	}
	if err := cfg.Load("config.yaml"); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to load config at config.yaml")
	}

	// Expand environment variables in the DSN
	db, err := dbpg.New(os.ExpandEnv(cfg.GetString("db.dsn")), nil, nil)
	if err != nil {
		zlog.Logger.Fatal().Msg("failed to connect to database")
	}

	st := postgres.New(db)
	srv := api.New(st)

	if err := srv.Run(cfg.GetString("server.addr")); err != nil {
		zlog.Logger.Fatal().Msg("failed to run server")
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	zlog.Logger.Info().Msg("shutting down server")
	if err := st.Close(); err != nil {
		zlog.Logger.Fatal().Msg("failed to close storage")
	}

	zlog.Logger.Info().Msg("server shutdown")
}
