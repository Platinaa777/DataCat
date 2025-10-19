package bootstrap

import (
	"context"
	"log/slog"
	"os"

	"github.com/Platinaa777/DataCat/internal/config"
)

func Run(ctx context.Context, migrate, migrateOnly bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	app, err := newApplication(ctx, cfg, log, migrate || migrateOnly)
	if err != nil {
		return err
	}
	defer app.close()
	if migrateOnly {
		return nil
	}
	if err := app.configureHTTP(ctx, cfg); err != nil {
		return err
	}
	return app.run(ctx)
}
