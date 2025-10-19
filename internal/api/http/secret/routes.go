package secret

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "POST", "/secret/add", s.Put)
	s.responses.Route(r, "GET", "/secret/{key}", s.Get)
	s.responses.Route(r, "DELETE", "/secret/remove/{key}", s.Delete)
}
