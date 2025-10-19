package bootstrap

import (
	"context"
	"net/http"
	"time"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/config"
)

func (a *application) configureHTTP(ctx context.Context, cfg config.Config) error {
	startup, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	identity, err := auth.New(startup, cfg, a.service)
	if err != nil {
		return err
	}
	router := NewRouter(a.service, identity, a.provider, cfg, a.log, a.cache)
	a.http = &http.Server{Addr: cfg.HTTPAddr, Handler: router.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 35 * time.Second, IdleTimeout: 90 * time.Second, MaxHeaderBytes: 16 << 10}
	return nil
}
