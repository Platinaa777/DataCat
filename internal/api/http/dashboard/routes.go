package dashboard

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "POST", "/dashboard/add", s.Create)
	s.responses.Route(r, "GET", "/dashboard/{id}", s.Get)
	s.responses.Route(r, "GET", "/dashboard/{id}/full", s.Full)
	s.responses.Route(r, "POST", "/dashboard/search", s.Search)
	s.responses.Route(r, "PUT", "/dashboard/update/{id}", s.Update)
	s.responses.Route(r, "DELETE", "/dashboard/remove/{id}", s.Delete)
}
