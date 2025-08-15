package main

import "github.com/wb-go/wbf/zlog"

func main() {
	zlog.Init()
	log := zlog.Logger
	log.Debug().Msg("debug enabled")

}
