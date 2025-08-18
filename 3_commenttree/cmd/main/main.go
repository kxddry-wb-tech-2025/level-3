package main

import (
	"os"

	"github.com/kxddry/wbf/config"
	"github.com/kxddry/wbf/zlog"
	"github.com/subosito/gotenv"
)

func main() {
	gotenv.Load(".env")
	zlog.Init()
	cfg := config.New()
	_ = cfg.Load(os.Getenv("CONFIG_PATH"))
	_ = cfg.Load("config.yaml")
}
