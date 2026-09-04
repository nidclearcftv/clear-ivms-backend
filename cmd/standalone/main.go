package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nidclearcftv/clear-ivms-backend/adapter/cmsv6"
	"github.com/nidclearcftv/clear-ivms-backend/adapter/db/postgres"
	httpapi "github.com/nidclearcftv/clear-ivms-backend/adapter/http"
	"github.com/nidclearcftv/clear-ivms-backend/core/service"
	"github.com/nidclearcftv/clear-ivms-backend/utils/env"
	"github.com/nidclearcftv/clear-ivms-backend/utils/logger"
)

type Env struct {
	DatabaseURL            string `env:"DATABASE_URL,required=true"`
	DatabaseSchemaPath     string `env:"DATABASE_SCHEMA_PATH,default=adapter/db/postgres/sql/schema.sql"`
	DatabaseMigrationsPath string `env:"DATABASE_MIGRATIONS_PATH,default=adapter/db/postgres/sql/migrations"`

	HTTPAddr           string   `env:"HTTP_ADDR,default=:8080"`
	HTTPAllowedOrigins []string `env:"HTTP_ALLOWED_ORIGINS,separator=,"`
}

type App struct {
	CMSV6 *cmsv6.Server
	HTTP  *httpapi.Server
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log, err := logger.New(logger.Options{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer log.Sync()

	var envOptions Env
	err = env.LoadEnv(&envOptions)
	if err != nil {
		log.Fatalw("failed to load environment configuration", "error", err)
	}

	db, err := postgres.NewDB(ctx, postgres.Options{
		Logger:           log,
		ConnectionString: envOptions.DatabaseURL,
		SchemaPath:       envOptions.DatabaseSchemaPath,
		MigrationsPath:   envOptions.DatabaseMigrationsPath,
	})
	if err != nil {
		log.Fatalw("failed to create database connection", "error", err)
	}
	defer db.Close()

	if err := db.Initialize(ctx); err != nil {
		log.Fatalw("failed to initialize database", "error", err)
	}

	vehicleRepository := postgres.NewVehicleRepository(db)

	vehicleService, err := service.NewVehicleService(service.VehicleServiceOptions{
		Repository: vehicleRepository,
	})
	if err != nil {
		log.Fatalw("failed to create vehicle service", "error", err)
	}

	httpServer, err := httpapi.NewServer(httpapi.Options{
		Logger:         log,
		Addr:           envOptions.HTTPAddr,
		AllowedOrigins: envOptions.HTTPAllowedOrigins,
		VehicleService: vehicleService,
	})
	if err != nil {
		log.Fatalw("failed to create http server", "error", err)
	}
	defer httpServer.Close()

	if err := httpServer.Start(); err != nil {
		log.Fatalw("failed to start http server", "error", err)
	}

	log.Info("clear-ivms-backend started")

	<-ctx.Done()

	log.Info("shutting down")
}
