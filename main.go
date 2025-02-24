package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hard-gainer/auth-service/internal/app"
	"github.com/hard-gainer/auth-service/internal/config"
)

func main() {
	cfg := config.MustLoad()
	log := initLogger()

	app := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)
	go func() {
		app.GRPCServer.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	app.GRPCServer.Stop()
	log.Info("Gracefully stopped")
}

func initLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}
