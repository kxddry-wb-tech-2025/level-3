package main

import (
	"delayed-notifier/internal/broker/rabbitmq"
	"os"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/ginext"
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

	engine := ginext.New()
	if err := engine.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
		os.Exit(1)
	}

}
