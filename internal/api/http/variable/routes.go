package variable

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "POST", "/variable/add", s.Create)
	s.responses.Route(r, "GET", "/variable/{id}", s.Get)
	s.responses.Route(r, "GET", "/variables/dashboard/{id}", s.ByDashboard)
	s.responses.Route(r, "GET", "/variables/namespace/{id}", s.ByNamespace)

	s.responses.Route(r, "DELETE", "/variable/remove/{id}", s.Delete)
	s.responses.Route(r, "PUT", "/variable/update/{id}", s.Update)
}
