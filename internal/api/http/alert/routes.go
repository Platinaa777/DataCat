package alert

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "POST", "/alert/add", s.Create)
	s.responses.Route(r, "GET", "/alert/get-counters", s.Counters)
	s.responses.Route(r, "GET", "/alert/template-vars", s.TemplateVariables)
	s.responses.Route(r, "GET", "/alert/{id}", s.Get)
	s.responses.Route(r, "POST", "/alert/search", s.Search)
	s.responses.Route(r, "DELETE", "/alert/remove/{id}", s.Delete)
	s.responses.Route(r, "PUT", "/alert/update/{id}", s.Update)
	s.responses.Route(r, "POST", "/alert/mute/{id}", s.Mute)
}
