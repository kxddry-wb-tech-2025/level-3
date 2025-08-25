package main

import (
	"os"
	"warehousecontrol/src/internal/config"
	"warehousecontrol/src/internal/delivery"

	"github.com/kxddry/wbf/zlog"
)

func main() {
	cfg := config.MustLoad(os.Getenv("CONFIG_PATH"))
	zlog.Init()

	server := delivery.NewServer(cfg)

	server.Run(cfg.Server.Address)
}
