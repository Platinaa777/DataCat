package datasource

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "POST", "/data-source/add", s.Create)
	s.responses.Route(r, "GET", "/data-source/{id}", s.Get)
	s.responses.Route(r, "POST", "/data-source/search", s.Search)
	s.responses.Route(r, "PUT", "/data-source/update-connection-string/{id}", s.UpdateConnection)
	s.responses.Route(r, "DELETE", "/data-source/remove/{id}", s.Delete)
}
