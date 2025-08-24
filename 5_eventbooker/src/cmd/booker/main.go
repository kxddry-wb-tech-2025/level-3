package main

import (
	"context"
	"eventbooker/src/internal/api/client"
	"eventbooker/src/internal/api/kafka"
	api "eventbooker/src/internal/api/server"
	"eventbooker/src/internal/config"
	"eventbooker/src/internal/domain/usecase"
	"eventbooker/src/internal/storage/txmanager"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kxddry/wbf/zlog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.MustLoadConfig(os.Getenv("CONFIG_FILE"))
	zlog.Init()
	log := zlog.Logger
	txmgr, err := txmanager.New(cfg.Storage.MasterDSN, cfg.Storage.SlaveDSNs...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create tx manager")
	}

	cli := client.New(cfg.NotifAddr)
	kfk, err := kafka.NewConsumer(ctx, cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka consumer")
	}
	uc := usecase.New(ctx, kfk, cli, txmgr)
	log.Info().Msg("usecase created")

	srv := api.NewServer(uc)
	if err := srv.Run(cfg.Srv.Addrs...); err != nil {
		log.Error().Err(err).Msg("failed to run server")
		cancel()
		time.Sleep(10 * time.Second) // wait for other goroutines to finish
	}
	log.Info().Msg("server started")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
	log.Info().Msg("server stopped")

	if err := txmgr.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close tx manager")
	}
	log.Info().Msg("tx manager closed")

	if err := kfk.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close kafka consumer")
	}
	log.Info().Msg("kafka consumer closed")

	log.Info().Msg("bye")
}
