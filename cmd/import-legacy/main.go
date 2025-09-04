package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/Platinaa777/DataCat/internal/pkg/secrets"
	"github.com/Platinaa777/DataCat/internal/repository/legacy"
	"github.com/Platinaa777/DataCat/internal/repository/postgres"
)

func main() {
	if err := run(); err != nil {
		slog.Error("legacy import failed", "error", err)
		os.Exit(1)
	}
}
func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	key, err := base64.StdEncoding.DecodeString(os.Getenv("SECRET_KEY"))
	if err != nil {
		return err
	}
	cipher, err := secrets.New(key)
	if err != nil {
		return err
	}
	store, err := postgres.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer store.Close()
	if err = store.Migrate(ctx); err != nil {
		return err
	}
	counts, err := legacy.Import(ctx, os.Getenv("LEGACY_DATABASE_URL"), store, cipher)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(counts)
}
