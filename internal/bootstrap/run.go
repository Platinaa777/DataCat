package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	grpcapi "github.com/Platinaa777/DataCat/internal/api/grpc"
	httpapi "github.com/Platinaa777/DataCat/internal/api/http"
	"github.com/Platinaa777/DataCat/internal/config"
	"github.com/Platinaa777/DataCat/internal/extapi/httpclient"
	"github.com/Platinaa777/DataCat/internal/extapi/notifications"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/extapi/vault"
	"github.com/Platinaa777/DataCat/internal/pkg/cache"
	"github.com/Platinaa777/DataCat/internal/pkg/secrets"
	"github.com/Platinaa777/DataCat/internal/repository/postgres"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/Platinaa777/DataCat/internal/worker"
)

func Run(ctx context.Context, migrate, migrateOnly bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	startup, cancel := context.WithTimeout(ctx, 30*time.Second)
	store, err := postgres.Open(startup, cfg.DatabaseURL)
	cancel()
	if err != nil {
		return err
	}
	defer store.Close()
	if migrate || migrateOnly {
		migrationCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err = store.Migrate(migrationCtx)
		cancel()
		if err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
	}
	cipher, err := secrets.New(cfg.SecretKey)
	if err != nil {
		return err
	}
	service := usecases.New(store, cipher)
	if err = service.Seed(ctx); err != nil {
		return fmt.Errorf("seed integration types: %w", err)
	}
	if migrateOnly {
		return nil
	}
	client := httpclient.New(cfg.AllowedHosts)
	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		if _, err = client.Validate(addr); err != nil {
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
	}
	var resultCache cache.Cache = cache.NewMemory(1000)
	if cfg.RedisURL != "" {
		redis, err := cache.NewRedis(cfg.RedisURL)
		if err != nil {
			return fmt.Errorf("Redis URL: %w", err)
		}
		defer redis.Close()
		resultCache = redis
	}
	identity, err := auth.New(ctx, cfg, service)
	if err != nil {
		return err
	}
	provider := &telemetry.Provider{Client: client}
	api := &httpapi.Server{Service: service, Auth: identity, Telemetry: provider, Config: cfg, Log: log, Cache: resultCache}
	httpServer := &http.Server{Addr: cfg.HTTPAddr, Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 35 * time.Second, IdleTimeout: 90 * time.Second, MaxHeaderBytes: 16 << 10}
	listener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen gRPC: %w", err)
	}
	defer listener.Close()
	grpcServer := grpcapi.New(identity, store)
	ctx, stop := context.WithCancel(ctx)
	defer stop()
	worker := &worker.Worker{Store: store, Service: service, Telemetry: provider, Sender: &notifications.Sender{Client: client, Service: service}, Log: log}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); worker.Run(ctx) }()
	errorsCh := make(chan error, 2)
	go func() { errorsCh <- httpServer.ListenAndServe() }()
	go func() { errorsCh <- grpcServer.Serve(listener) }()
	log.Info("server started", "http_addr", cfg.HTTPAddr, "grpc_addr", cfg.GRPCAddr)
	select {
	case <-ctx.Done():
		err = nil
	case serveErr := <-errorsCh:
		if !errors.Is(serveErr, http.ErrServerClosed) {
			err = serveErr
		}
	}
	stop()
	shutdown, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if shutdownErr := httpServer.Shutdown(shutdown); shutdownErr != nil {
		log.Warn("HTTP shutdown timed out")
		httpServer.Close()
	}
	grpcDone := make(chan struct{})
	go func() { grpcServer.GracefulStop(); close(grpcDone) }()
	select {
	case <-grpcDone:
	case <-shutdown.Done():
		grpcServer.Stop()
	}
	wg.Wait()
	log.Info("server stopped")
	return err
}
