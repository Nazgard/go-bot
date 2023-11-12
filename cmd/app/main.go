package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"makarov.dev/bot/internal/background"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/delivery/web"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := log.New()
	config.Init(logger)

	background.StartAllBackgroundJobs(ctx)
	go web.StartWeb(ctx)

	log.Infof("Application started")

	<-ctx.Done()
	log.Infof("Gracefully shutdown application")
}
