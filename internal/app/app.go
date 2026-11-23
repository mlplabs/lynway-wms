package app

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/mlplabs/common-go-pkg/pkg/db/postgres"
	httpServer "github.com/mlplabs/common-go-pkg/pkg/http/server"
	"github.com/mlplabs/lynway-wms/internal/config"
	controller "github.com/mlplabs/lynway-wms/internal/controller/http"
	v1 "github.com/mlplabs/lynway-wms/internal/controller/http/v1"
	"github.com/mlplabs/lynway-wms/internal/repository"
	"github.com/mlplabs/lynway-wms/internal/repository/bootstrap"
	"github.com/mlplabs/lynway-wms/internal/repository/print"
	"github.com/mlplabs/lynway-wms/internal/repository/users"
	"github.com/mlplabs/lynway-wms/internal/usecases/wms"
	authClient "gitlab.com/mlplabs/auth-service/pkg/client"
)

const applicationName = "lynway-wms"

// Run запускает сервис.
func Run(cfg *config.Config) {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	pgConfig := postgres.Config{
		PoolMax:  cfg.PgDb.PoolMax,
		Host:     cfg.PgDb.Host,
		Port:     cfg.PgDb.Port,
		User:     cfg.PgDb.User,
		Password: cfg.PgDb.Password,
		DB:       cfg.PgDb.DB,
	}

	pgConnection, err := postgres.Connect(&pgConfig)
	if err != nil {
		log.Printf("failed to connect to Postgres DB: %s", err)
		return
	}
	defer pgConnection.Close()

	err = bootstrap.RunMigrations(pgConnection)
	if err != nil {
		log.Printf("failed to run migrations: %s", err)
		return
	}

	err = bootstrap.SeedDefaults(ctx, pgConnection)
	if err != nil {
		log.Printf("failed to seed defaults: %s", err)
		return
	}

	service := wms.NewService(
		authClient.NewHttpClient(applicationName, &cfg.AuthClient),
		repository.NewWhsRepository(pgConnection),
		users.NewUsersRepository(pgConnection),
		print.NewPrintRepository(pgConnection),
	)

	router := controller.NewRouter()
	apiRouter := chi.NewRouter()

	v1.NewServiceHandler(apiRouter, nil, service)

	router.Mount("/api", apiRouter)

	server := httpServer.NewServer(&router, httpServer.Port(cfg.HTTP.Port))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM)

	log.Println("app.Run: service successful started")

	select {
	case s := <-interrupt:
		log.Printf("app.Run: terminated, signal: %s", s.String())
	case err = <-server.Notify():
		log.Printf("app.Run: httpServer.Notify: %s", err)
		return
	}

	err = server.Shutdown()
	if err != nil {
		log.Printf("app.Run: httpServer.Shutdown: %s", err)
		return
	}
}
