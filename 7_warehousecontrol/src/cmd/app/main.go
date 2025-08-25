package main

import (
	"os"
	"warehousecontrol/src/internal/config"
	"warehousecontrol/src/internal/delivery"
	historyrepo "warehousecontrol/src/internal/repo/history"
	itemrepo "warehousecontrol/src/internal/repo/items"
	"warehousecontrol/src/internal/repo/users"
	"warehousecontrol/src/internal/service/auth"
	historyuc "warehousecontrol/src/internal/service/history"
	itemuc "warehousecontrol/src/internal/service/items"

	"github.com/kxddry/wbf/zlog"
)

func main() {
	cfg := config.MustLoad(os.Getenv("CONFIG_PATH"))
	zlog.Init()
	_ = cfg

	userRepo, err := users.New(cfg.Storage.MasterDSN, cfg.Storage.SlaveDSNs...)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to create user repository")
	}

	itemRepo, err := itemrepo.New(cfg.Storage.MasterDSN, cfg.Storage.SlaveDSNs...)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to create item repository")
	}

	historyRepo, err := historyrepo.New(cfg.Storage.MasterDSN, cfg.Storage.SlaveDSNs...)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to create history repository")
	}

	userUsecase := auth.NewUsecase(userRepo, cfg.JWT.Secret)
	itemUsecase := itemuc.NewUsecase(itemRepo)
	historyUsecase := historyuc.NewUsecase(historyRepo)
	srv := delivery.NewServer(cfg, itemUsecase, userUsecase, historyUsecase)

	if err := srv.Run(cfg.Server.Address); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to run server")
	}

	// no graceful shutdown here because we don't have any long-running processes
	userRepo.Close()
	itemRepo.Close()
	historyRepo.Close()

}
