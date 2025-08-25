package main

import (
	"os"
	"salestracker/src/internal/config"
	"salestracker/src/internal/delivery"
	"salestracker/src/internal/service"
	"salestracker/src/internal/storage/txmanager"

	"github.com/joho/godotenv"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	godotenv.Load()
	cfg, err := config.New(os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}
	cfg.Storage.Password = os.ExpandEnv(cfg.Storage.Password)

	zlog.Init()

	txmgr, err := txmanager.New(cfg.GetMasterDSN())
	if err != nil {
		panic(err)
	}
	defer txmgr.Close()
	uc := service.NewUsecase(txmgr)

	srv := delivery.New(uc, cfg.Server.StaticDir)
	if err := srv.Run(cfg.Server.Addrs...); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to run server")
	}

	os.Exit(0)
}
