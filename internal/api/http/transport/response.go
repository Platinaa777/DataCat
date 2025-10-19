package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Endpoint func(*http.Request) (any, error)
type Responder struct {
	Log      *slog.Logger
	Requests atomic.Uint64
	Failed   atomic.Uint64
}

func (s *Responder) Route(r chi.Router, method, path string, fn Endpoint) {
	r.MethodFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		start := time.Now()
		result, err := fn(r)
		if err != nil {
			s.Failed.Add(1)
			s.Log.WarnContext(ctx, "request failed", "method", r.Method, "path", r.URL.Path, "request_id", middleware.GetReqID(ctx), "error", err)
			Problem(w, r, err)
			return
		}
		Write(w, http.StatusOK, result)
		s.Log.DebugContext(ctx, "request completed", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

func Write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if value != nil {
		json.NewEncoder(w).Encode(value)
	}
}

func Problem(w http.ResponseWriter, r *http.Request, err error) {
	status := 500
	detail := "Request failed"
	switch {
	case errors.Is(err, entity.ErrInvalid):
		status = 400
		detail = err.Error()
	case errors.Is(err, entity.ErrForbidden):
		status = 403
		detail = "Access denied"
	case errors.Is(err, entity.ErrNotFound):
		status = 404
		detail = "Resource not found"
	case errors.Is(err, entity.ErrConflict):
		status = 409
		detail = "Resource already exists, is in use, or was modified"
	case errors.Is(err, entity.ErrUpstream):
		status = 502
		detail = "Telemetry provider unavailable"
	case errors.Is(err, context.DeadlineExceeded):
		status = 504
		detail = "Request timed out"
	}
	Write(w, status, map[string]any{"type": "about:blank", "title": http.StatusText(status), "status": status, "detail": detail, "instance": r.URL.Path, "errors": map[string][]string{"General": {detail}}, "requestId": middleware.GetReqID(r.Context())})
}

func Upstream(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", entity.ErrUpstream, err)
}
