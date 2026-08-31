package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nidclearcftv/clear-ivms-backend/adapter/cmsv6"
	"github.com/nidclearcftv/clear-ivms-backend/utils/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log, err := logger.New(logger.Options{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer log.Sync()

	cfg, err := loadEnv()
	if err != nil {
		log.Fatalw("failed to load environment configuration", "error", err)
	}

	server, err := cmsv6.NewServer(cmsv6.Options{
		Context:  ctx,
		Logger:   log,
		URL:      cfg.CMSV6URL,
		Username: cfg.CMSV6Username,
		Password: cfg.CMSV6Password,
	})
	if err != nil {
		log.Fatalw("failed to create cmsv6 server", "error", err)
	}
	defer server.Close()

	log.Info("clear-ivms-backend started")

	err = server.Start()
	if err != nil {
		log.Fatalw("failed to start cmsv6 server", "error", err)
	}

	<-ctx.Done()

	log.Info("shutting down")
}
