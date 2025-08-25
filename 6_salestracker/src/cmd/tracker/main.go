package main

import (
	"context"
	"os"
	"salestracker/src/internal/config"

	"github.com/wb-go/wbf/zlog"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.New(os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}

	zlog.Init()
	_ = ctx
	_ = cfg
}
