package main

import (
	"context"
	"image-processor/internal/api"
	"image-processor/internal/broker/kafka"
	"image-processor/internal/storage/minio"
	"image-processor/internal/storage/postgres"
	"image-processor/internal/usecase"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kxddry/wbf/config"
	"github.com/kxddry/wbf/retry"
	"github.com/kxddry/wbf/zlog"
)

func main() {
	godotenv.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zlog.Init()
	cfg := config.New()
	strat := retry.Strategy{
		Attempts: 5,
		Delay:    1 * time.Second,
		Backoff:  2,
	}

	if err := cfg.Load(os.Getenv("CONFIG_PATH")); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to load config with CONFIG_PATH")
	}
	if err := cfg.Load("config.yaml"); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to load config with config.yaml")
	}

	prod := kafka.NewProducer([]string{cfg.GetString("kafka.brokers")}, "uploaded", strat)
	s3, err := minio.New(ctx, minio.Config{
		Endpoint:   cfg.GetString("s3.endpoint"),
		AccessKey:  os.Getenv("S3_ACCESS_KEY"),
		SecretKey:  os.Getenv("S3_SECRET_KEY"),
		BucketName: cfg.GetString("s3.bucket"),
		BaseURL:    cfg.GetString("s3.base_url"),
		SSL:        false, // not for production
	})
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to create minio client")
	}

	st, err := postgres.New(os.ExpandEnv(cfg.GetString("postgres.masterdsn")))
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to create postgres client")
	}

	h := usecase.New(s3, st, prod)

	srv := api.New(h)

	if err := srv.Run(cfg.GetString("server.addr")); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to run server")
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch

	if err := st.Close(); err != nil {
		zlog.Logger.Err(err).Msg("failed to close postgres client")
	}
	_ = prod.Close()

}
