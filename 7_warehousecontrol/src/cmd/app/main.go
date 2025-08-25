package main

import (
	"os"
	"warehousecontrol/src/internal/config"

	"github.com/kxddry/wbf/zlog"
)

func main() {
	cfg := config.MustLoad(os.Getenv("CONFIG_PATH"))
	zlog.Init()
	_ = cfg
}
