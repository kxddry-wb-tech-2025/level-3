package main

import (
	"context"
	"delayed-notifier/internal/httpapi"
	"delayed-notifier/internal/queue"
	"delayed-notifier/internal/sender"
	"delayed-notifier/internal/storage"
	"delayed-notifier/internal/worker"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kxddry/wbf/config"
	"github.com/kxddry/wbf/ginext"
	"github.com/kxddry/wbf/zlog"
	"github.com/subosito/gotenv"
)

func main() {
	zlog.Init()
	log := &zlog.Logger

	cfg := config.New()
	// Load .env if present to populate process environment
	_ = gotenv.Load(".env")
	if err := cfg.Load("./config.yaml"); err != nil {
		panic(err)
	}

	// Build server address from host and port
	host := cfg.GetString("server.host")
	if host == "" {
		host = "0.0.0.0" // default: this host
	}
	portStr := cfg.GetString("server.port")
	if portStr == "" {
		portStr = "8080" // default: 8080
	}
	if _, err := strconv.Atoi(portStr); err != nil {
		portStr = "8080" // default: 8080
	}
	addr := fmt.Sprintf("%s:%s", host, portStr)

	// Dependencies: Redis storage, RabbitMQ queue, Telegram sender
	redisPort := cfg.GetString("redis.port")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisDBStr := cfg.GetString("redis.db")
	if redisDBStr == "" {
		redisDBStr = "0"
	}
	redisDB, _ := strconv.Atoi(redisDBStr)
	redisCfg := storage.RedisConfig{
		Addr:     fmt.Sprintf("%s:%s", cfg.GetString("redis.host"), redisPort),
		Password: cfg.GetString("redis.password"),
		DB:       redisDB,
	}
	redisStore, err := storage.NewRedisStorage(context.Background(), redisCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init redis storage")
	}

	rabbitPort := cfg.GetString("rabbitmq.port")
	if rabbitPort == "" {
		rabbitPort = "5672"
	}
	rp, _ := strconv.Atoi(rabbitPort)
	rmqCfg := queue.RabbitConfig{
		Host:      cfg.GetString("rabbitmq.host"),
		Port:      rp,
		Username:  cfg.GetString("rabbitmq.username"),
		Password:  cfg.GetString("rabbitmq.password"),
		QueueName: cfg.GetString("rabbitmq.queue_name"),
	}
	rmq, err := queue.NewRabbit(rmqCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init rabbitmq")
	}

	tgTimeoutStr := cfg.GetString("telegram.timeout")
	if tgTimeoutStr == "" {
		tgTimeoutStr = "30"
	}
	tgTimeoutSec, _ := strconv.Atoi(tgTimeoutStr)
	// Prefer env var; fallback to config for local setups
	token := os.Getenv("TELEGRAM_API_TOKEN")
	if token == "" {
		token = cfg.GetString("telegram.bot_token")
	}
	telegram := sender.NewTelegramSender(
		token,
		time.Duration(tgTimeoutSec)*time.Second,
	)

	// Workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := worker.NewScheduler(redisStore, rmq)
	consumer := worker.NewConsumer(redisStore, rmq, telegram)

	// Start workers
	go scheduler.Run(ctx)
	go consumer.Run(ctx)

	// HTTP server
	r := ginext.New()

	staticDir := cfg.GetString("server.static_dir")
	if staticDir != "" {
		// Serve static UI
		httpapi.ServeStatic(r, "/", staticDir)
	}

	httpapi.RegisterRoutes(r, redisStore)

	// HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Info().Msgf("server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Err(err).Msg("failed to start server")
		}
	}()

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Info().Msg("shutdown signal received, shutting down...")

	// Stop accepting new HTTP requests and wait for in-flight to finish
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Err(err).Msg("http server shutdown error")
	}

	// Stop background workers
	cancel()

	// Close external resources
	if err := rmq.Close(); err != nil {
		log.Err(err).Msg("failed to close rabbitmq")
	}
	if err := redisStore.Close(); err != nil {
		log.Err(err).Msg("failed to close redis")
	}

	log.Info().Msg("shutdown complete")
}
