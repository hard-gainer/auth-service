package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/hard-gainer/auth-service/internal/app/grpc"
	"github.com/hard-gainer/auth-service/internal/db"
	auth "github.com/hard-gainer/auth-service/internal/service"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	dbUrl string,
	tokenTTL time.Duration,
) *App {
	storage, err := db.New(dbUrl)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
