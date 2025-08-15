package main

import (
	"context"
	"delayed-notifier/internal/broker/rabbitmq"
	"os"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	zlog.Init()
	log := zlog.Logger
	log.Debug().Msg("debug enabled")

	cfg := config.New()
	err := cfg.Load("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config.yaml")
		os.Exit(1)
	}

	rmqDsn := cfg.GetString("rabbitmq.dsn")
	if rmqDsn == "" {
		log.Fatal().Msg("rabbitmq dsn is empty")
		os.Exit(1)
	}

	rmq, err := rabbitmq.New(rmqDsn, "notify")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to rabbitmq")
		os.Exit(1)
	}

	_ = rmq

	addr := cfg.GetString("server.address")
	log.Debug().Str("address", addr).Msg("Starting server")

	psqlOpts := &dbpg.Options{
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	}

	db, err := dbpg.New(cfg.GetString("storage.dsn"), nil, psqlOpts)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
		os.Exit(1)
	}

	engine := ginext.New()
	_ = engine.SetTrustedProxies(nil) // disable warning

	if err := engine.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
		os.Exit(1)
	}

}
