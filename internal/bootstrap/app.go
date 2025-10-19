package bootstrap

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/pkg/cache"
	"github.com/Platinaa777/DataCat/internal/repository/postgres"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/Platinaa777/DataCat/internal/worker"
)

type application struct {
	store       *postgres.Store
	service     *usecases.Service
	provider    *telemetry.Provider
	cache       cache.Cache
	cacheCloser io.Closer
	http        *http.Server
	worker      *worker.Worker
	log         *slog.Logger
}

func (a *application) close() {
	if a.cacheCloser != nil {
		a.cacheCloser.Close()
	}
	a.store.Close()
}

func (a *application) run(ctx context.Context) error {
	ctx, stop := context.WithCancel(ctx)
	defer stop()
	var workers sync.WaitGroup
	workers.Add(1)
	go func() { defer workers.Done(); a.worker.Run(ctx) }()
	result := make(chan error, 1)
	go func() { result <- a.http.ListenAndServe() }()
	a.log.Info("server started", "http_addr", a.http.Addr)
	var err error
	select {
	case <-ctx.Done():
	case err = <-result:
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
	}
	stop()
	shutdown, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if shutdownErr := a.http.Shutdown(shutdown); shutdownErr != nil {
		a.log.Warn("HTTP shutdown timed out")
		a.http.Close()
	}
	workers.Wait()
	a.log.Info("server stopped")
	return err
}
