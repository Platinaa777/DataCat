package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Platinaa777/DataCat/internal/config"
	"github.com/Platinaa777/DataCat/internal/extapi/httpclient"
	"github.com/Platinaa777/DataCat/internal/extapi/notifications"
	"github.com/Platinaa777/DataCat/internal/extapi/plugins"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/extapi/vault"
	"github.com/Platinaa777/DataCat/internal/pkg/cache"
	"github.com/Platinaa777/DataCat/internal/pkg/secrets"
	"github.com/Platinaa777/DataCat/internal/repository/postgres"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/Platinaa777/DataCat/internal/worker"
)

func newApplication(ctx context.Context, cfg config.Config, log *slog.Logger, migrate bool) (_ *application, err error) {
	startup, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	store, err := postgres.Open(startup, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	app := &application{store: store, log: log}
	defer func() {
		if err != nil {
			app.close()
		}
	}()
	if migrate {
		if err = store.Migrate(startup); err != nil {
			return nil, fmt.Errorf("migrate database: %w", err)
		}
	}
	cipher, err := secrets.New(cfg.SecretKey)
	if err != nil {
		return nil, err
	}
	client := httpclient.New(cfg.AllowedHosts)
	app.service = usecases.New(store, cipher, client)
	if err = app.service.Seed(startup); err != nil {
		return nil, fmt.Errorf("seed integration types: %w", err)
	}
	if err = configureVault(app.service, client); err != nil {
		return nil, err
	}
	var resultCache cache.Cache = cache.NewMemory(1000)
	if cfg.RedisURL != "" {
		redis, openErr := cache.NewRedis(cfg.RedisURL)
		if openErr != nil {
			return nil, fmt.Errorf("Redis URL: %w", openErr)
		}
		resultCache, app.cacheCloser = redis, redis
	}
	provider := &telemetry.Provider{Client: client}
	app.queries = usecases.NewQueries(app.service, provider, &plugins.Client{HTTP: client}, resultCache)
	app.worker = &worker.Worker{Store: store, Service: app.service, Telemetry: provider, Sender: &notifications.Sender{Client: client, Service: app.service}, Log: log}
	return app, nil
}

func configureVault(service *usecases.Service, client *httpclient.Client) error {
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		return nil
	}
	if _, err := client.Validate(addr); err != nil {
		return fmt.Errorf("Vault address: %w", err)
	}
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		return fmt.Errorf("VAULT_TOKEN is required")
	}
	mount := os.Getenv("VAULT_MOUNT")
	if mount == "" {
		mount = "secret"
	}
	service.RemoteSecrets = &vault.Vault{Client: client, Address: addr, Token: token, Mount: mount}
	return nil
}
