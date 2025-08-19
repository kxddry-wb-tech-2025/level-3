package main

import (
	"context"
	"image-processor/internal/broker/kafka"
	"image-processor/internal/domain"
	editworker "image-processor/internal/edit-worker"
	"image-processor/internal/editor"
	"image-processor/internal/storage/minio"
	"image-processor/internal/storage/postgres"
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

	st, err := postgres.New(os.ExpandEnv(cfg.GetString("postgres.masterdsn")))
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to create postgres client")
	}

	s3, err := minio.New(ctx, minio.Config{
		Endpoint:   cfg.GetString("s3.endpoint"),
		AccessKey:  os.Getenv("S3_ACCESS_KEY"),
		SecretKey:  os.Getenv("S3_SECRET_KEY"),
		BucketName: cfg.GetString("s3.bucket"),
	})

	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to create s3 client")
	}

	// nothing better to do with this SHITTY AHH config package
	cons := kafka.NewConsumer([]string{cfg.GetString("kafka.brokers")}, "uploaded", "resizer", strat)
	ch := make(chan *domain.KafkaMessage)
	cons.StartConsuming(ctx, ch)

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
	<-shutdownCh

	worker := editworker.NewWorker(editor.NewEditor(), st, s3)
	go worker.Handle(ctx, ch)

	if err := st.Close(); err != nil {
		zlog.Logger.Err(err).Msg("failed to close postgres client")
	}
	_ = cons.Close()
}
