package panel

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "GET", "/panel/types", s.Types)
	s.responses.Route(r, "POST", "/panel/add", s.Create)
	s.responses.Route(r, "GET", "/panel/{id}", s.Get)
	s.responses.Route(r, "DELETE", "/panel/remove/{id}", s.Delete)
	s.responses.Route(r, "PUT", "/panel/update/{id}", s.Update)
}
