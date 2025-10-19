package plugin

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "POST", "/plugin/add", s.Create)
	s.responses.Route(r, "GET", "/plugin/search", s.Search)
	s.responses.Route(r, "GET", "/plugin/{id}", s.Get)
	s.responses.Route(r, "DELETE", "/plugin/remove/{id}", s.Delete)
	s.responses.Route(r, "POST", "/plugin/{id}", s.SetActive)
	s.responses.Route(r, "PUT", "/plugin/update/{id}", s.Update)
}
