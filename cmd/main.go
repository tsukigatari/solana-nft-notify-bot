package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ndfz/solana-nft-notify-bot/internal/config"
	"github.com/ndfz/solana-nft-notify-bot/internal/magiceden"
	"github.com/ndfz/solana-nft-notify-bot/internal/magiceden/worker"
	"github.com/ndfz/solana-nft-notify-bot/internal/services"
	"github.com/ndfz/solana-nft-notify-bot/internal/storage"
	"go.uber.org/zap"
)

func main() {
	config, err := config.New()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger := zap.Must(zap.NewProduction())
	if os.Getenv("APP_ENV") == "development" {
		logger = zap.Must(zap.NewDevelopment())
		logger.Debug("development mode")
	}
	defer logger.Sync()

	zap.ReplaceGlobals(logger)
	zap.S().Info("starting solana-nft-notify-bot")

	storage, err := storage.New(config.DatabaseName)
	if err != nil {
		panic("failed to open database: " + err.Error())
	}
	defer storage.DB.Close()
	zap.S().Info("connected to SQLite database")
	err = storage.CreateTables()
	if err != nil {
		panic("failed to create tables: " + err.Error())
	}
	zap.S().Info("created tables")

	magiceden := magiceden.New(config.MagicEdenEndpoint)

	services := services.New(
		config,
		storage,
		magiceden,
	)

	go worker.New(services).Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
}