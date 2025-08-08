package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/config"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/pkg/cache"
	"github.com/Platinaa777/DataCat/internal/usecases"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type Server struct {
	Service   *usecases.Service
	Auth      *auth.Auth
	Telemetry *telemetry.Provider
	Config    config.Config
	Log       *slog.Logger
	Cache     cache.Cache
	requests  atomic.Uint64
	failed    atomic.Uint64
}
type handler func(*http.Request) (any, error)

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer, s.headers)
	r.Get("/health/live", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if s.Service.Store.Ping(ctx) != nil {
			write(w, 503, map[string]string{"status": "unavailable"})
			return
		}
		write(w, 200, map[string]string{"status": "ok"})
	})
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# TYPE datacat_http_requests_total counter\ndatacat_http_requests_total %d\n# TYPE datacat_http_errors_total counter\ndatacat_http_errors_total %d\n", s.requests.Load(), s.failed.Load())
	})
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/user/login-code-flow", s.Auth.Start)
		r.Get("/user/callback", s.Auth.Callback)
		r.With(middleware.Throttle(10)).Post("/user/login", auth.LimitLogin(s.Auth.Password))
		r.Group(func(r chi.Router) { r.Use(s.Auth.Middleware); s.resources(r); s.integrations(r) })
	})
	return r
}
func (s *Server) headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.requests.Add(1)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		origin := r.Header.Get("Origin")
		if origin != "" {
			if origin != s.Config.AllowedOrigin {
				problem(w, r, entity.ErrForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Namespace, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" && origin == "" && r.Header.Get("Authorization") == "" {
			if _, err := r.Cookie("access_token"); err == nil {
				problem(w, r, entity.ErrForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) route(r chi.Router, method, path string, fn handler) {
	r.MethodFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		start := time.Now()
		result, err := fn(r)
		if err != nil {
			s.failed.Add(1)
			s.Log.WarnContext(ctx, "request failed", "method", r.Method, "path", r.URL.Path, "request_id", middleware.GetReqID(ctx), "error", err)
			problem(w, r, err)
			return
		}
		write(w, http.StatusOK, result)
		s.Log.DebugContext(ctx, "request completed", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
func write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if value != nil {
		json.NewEncoder(w).Encode(value)
	}
}
func problem(w http.ResponseWriter, r *http.Request, err error) {
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
	write(w, status, map[string]any{"type": "about:blank", "title": http.StatusText(status), "status": status, "detail": detail, "instance": r.URL.Path, "errors": map[string][]string{"General": {detail}}, "requestId": middleware.GetReqID(r.Context())})
}
func body[T any](r *http.Request) (T, error) {
	var value T
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&value); err != nil {
		return value, fmt.Errorf("%w: invalid JSON body", entity.ErrInvalid)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return value, fmt.Errorf("%w: only one JSON value is allowed", entity.ErrInvalid)
	}
	return value, nil
}
func (s *Server) scope(r *http.Request, write bool) (string, error) {
	ns := r.Header.Get("X-Namespace")
	if _, err := uuid.Parse(ns); err != nil {
		return "", fmt.Errorf("%w: X-Namespace must be a UUID", entity.ErrInvalid)
	}
	if !auth.Principal(r.Context()).Can(ns, write) {
		return "", entity.ErrForbidden
	}
	if _, err := s.Service.Store.Get(r.Context(), usecases.Global, "namespace", ns); err != nil {
		return "", err
	}
	return ns, nil
}
func admin(r *http.Request) error {
	if !auth.Principal(r.Context()).GlobalAdmin {
		return entity.ErrForbidden
	}
	return nil
}
func (s *Server) namespaceAdmin(r *http.Request) (string, error) {
	ns, err := s.scope(r, true)
	if err != nil {
		return "", err
	}
	p := auth.Principal(r.Context())
	if p.GlobalAdmin {
		return ns, nil
	}
	for _, c := range p.Claims {
		if c.NamespaceID == ns && c.Role == "Admin" {
			return ns, nil
		}
	}
	return "", entity.ErrForbidden
}
func param(r *http.Request, key string) string { return chi.URLParam(r, key) }
func pagination(r *http.Request) (int, int, error) {
	page, size := 1, 10
	var err error
	if v := r.URL.Query().Get("page"); v != "" {
		page, err = strconv.Atoi(v)
		if err != nil {
			return 0, 0, entity.ErrInvalid
		}
	}
	if v := r.URL.Query().Get("pageSize"); v != "" {
		size, err = strconv.Atoi(v)
		if err != nil {
			return 0, 0, entity.ErrInvalid
		}
	}
	return page, size, nil
}
func search[T any](s *Server, r *http.Request, kind string) (any, error) {
	ns, err := s.scope(r, false)
	if err != nil {
		return nil, err
	}
	query, err := body[entity.Search](r)
	if err != nil {
		return nil, err
	}
	page, size, err := pagination(r)
	if err != nil {
		return nil, err
	}
	return resourcePage[T](s, r, ns, kind, query, page, size)
}
func resourcePage[T any](s *Server, r *http.Request, ns, kind string, query entity.Search, page, size int) (entity.Page[T], error) {
	records, total, err := s.Service.Store.Search(r.Context(), ns, kind, query, page, size)
	if err != nil {
		return entity.Page[T]{}, err
	}
	out := make([]T, 0, len(records))
	for _, record := range records {
		var value T
		if err = json.Unmarshal(record.Data, &value); err != nil {
			return entity.Page[T]{}, err
		}
		out = append(out, value)
	}
	return entity.NewPage(out, total, page, size), nil
}

func get[T any](s *Server, r *http.Request, kind string) (any, error) {
	ns, err := s.scope(r, false)
	if err != nil {
		return nil, err
	}
	v, _, err := usecases.Read[T](r.Context(), s.Service.Store, ns, kind, param(r, "id"))
	return v, err
}
func (s *Server) remove(r *http.Request, kind string) (any, error) {
	ns, err := s.scope(r, true)
	if err != nil {
		return nil, err
	}
	return nil, s.Service.Delete(r.Context(), ns, kind, param(r, "id"))
}
func panelName(n int) string { return map[int]string{1: "Graph", 2: "Table", 3: "Pie Chart"}[n] }
func queryValue(r *http.Request, key string) string {
	for k, vs := range r.URL.Query() {
		if strings.EqualFold(k, key) && len(vs) > 0 {
			return vs[0]
		}
	}
	return ""
}
