package namespace

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "POST", "/namespace", s.Create)
	s.responses.Route(r, "GET", "/namespace/get-available", s.Available)
	s.responses.Route(r, "GET", "/namespace/{name}", s.Get)
	s.responses.Route(r, "POST", "/external-role", s.MapRole)
}
