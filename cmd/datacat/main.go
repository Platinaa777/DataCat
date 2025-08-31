package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Platinaa777/DataCat/internal/bootstrap"
)

func main() {
	migrate := flag.Bool("migrate", false, "apply database migrations before startup")
	migrateOnly := flag.Bool("migrate-only", false, "apply database migrations and exit")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := bootstrap.Run(ctx, *migrate, *migrateOnly); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
