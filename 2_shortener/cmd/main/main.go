package main

import (
	"time"

	"github.com/kxddry/wbf/config"
	"github.com/kxddry/wbf/dbpg"
	"github.com/kxddry/wbf/zlog"
)

func main() {
	zlog.Init()

	cfg := config.New()
	cfg.Load("./config.yaml")




}
