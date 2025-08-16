package main

import (
	"context"
	"delayed-notifier/internal/broker/rabbitmq"
	"delayed-notifier/internal/handlers"
	"delayed-notifier/internal/processor"
	"delayed-notifier/internal/sender/telegram"
	"delayed-notifier/internal/storage/redis"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kxddry/wbf/config"
	"github.com/kxddry/wbf/ginext"
	"github.com/kxddry/wbf/zlog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	zlog.Init()
	log := &zlog.Logger
	log.Debug().Msg("debug enabled")

	cfg := config.New()
	err := cfg.Load("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config.yaml")
		os.Exit(1)
	}

	rmq, err := rabbitmq.New(cfg.GetString("rabbitmq.dsn"), log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to rabbitmq")
		os.Exit(1)
	}

	r, err := redis.New(ctx, cfg.GetString("redis.address"), cfg.GetString("redis.password"), 0)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to redis")
		os.Exit(1)
	}
	_ = r

	timeout := cfg.GetString("telegram.timeout")
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse http.timeout")
		os.Exit(1)
	}
	bot, err := telegram.New(os.Getenv("TELEGRAM_API_TOKEN"), timeoutDuration)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to telegram")
		os.Exit(1)
	}

	np := processor.New(rmq, r, bot, log)
	np.Start(ctx)

	s := handlers.NewServer(rmq, r)

	engine := ginext.New()
	_ = engine.SetTrustedProxies(nil) // disable warning
	engine.LoadHTMLFiles("./static/index.html")

	// TODO: maintenance
	engine.GET("/notify/:id", s.GetNotification())
	engine.POST("/notify", s.PostNotification())
	engine.DELETE("/notify/:id", s.DeleteNotification())
	engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main website",
		})
	})

	addr := cfg.GetString("server.address")
	log.Debug().Str("address", addr).Msg("Starting server")
	if err := engine.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
		os.Exit(1)
	}

	// graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signalChan:
	case <-ctx.Done():
	}

	rmq.Close()

}
