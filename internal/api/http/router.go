package httpapi

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/auth"
	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Feature interface{ Register(chi.Router) }
type Server struct {
	Auth      *auth.Auth
	Config    config.Config
	Access    *transport.Access
	Responses *transport.Responder
	Features  []Feature
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer, s.headers)
	r.Get("/health/live", s.live)
	r.Get("/health/ready", s.ready)
	r.Get("/metrics", s.metrics)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/user/login-code-flow", s.Auth.Start)
		r.Get("/user/callback", s.Auth.Callback)
		r.With(middleware.Throttle(10)).Post("/user/login", auth.LimitLogin(s.Auth.Password))
		r.Group(func(r chi.Router) {
			r.Use(s.Auth.Middleware)
			for _, feature := range s.Features {
				feature.Register(r)
			}
		})
	})
	return r
}
