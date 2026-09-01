package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nidclearcftv/clear-ivms-backend/adapter/cmsv6"
	httpapi "github.com/nidclearcftv/clear-ivms-backend/adapter/http"
	"github.com/nidclearcftv/clear-ivms-backend/utils/env"
	"github.com/nidclearcftv/clear-ivms-backend/utils/logger"
)

type Env struct {
	CMSV6URL      string `env:"CMSV6_URL,required=true"`
	CMSV6Username string `env:"CMSV6_USERNAME,required=true"`
	CMSV6Password string `env:"CMSV6_PASSWORD,required=true"`

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

	cmsv6Server, err := cmsv6.NewServer(cmsv6.Options{
		Context:  ctx,
		Logger:   log,
		URL:      envOptions.CMSV6URL,
		Username: envOptions.CMSV6Username,
		Password: envOptions.CMSV6Password,
	})
	if err != nil {
		log.Fatalw("failed to create cmsv6 server", "error", err)
	}
	defer cmsv6Server.Close()

	if err := cmsv6Server.Start(); err != nil {
		log.Fatalw("failed to start cmsv6 server", "error", err)
	}

	httpServer, err := httpapi.NewServer(httpapi.Options{
		Logger:         log,
		Addr:           envOptions.HTTPAddr,
		AllowedOrigins: envOptions.HTTPAllowedOrigins,
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
