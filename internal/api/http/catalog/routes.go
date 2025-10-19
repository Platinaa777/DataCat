package catalog

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "POST", "/"+s.kind+"/add", s.Create)
	s.responses.Route(r, "GET", "/"+s.kind+"/get-all", s.List)
	s.responses.Route(r, "GET", "/"+s.kind+"/{name}", s.Get)
	path := "/" + s.kind + "/remove/{name}"
	if s.kind == "data-source-type" {
		path = "/datasource/remove/{name}"
	}
	s.responses.Route(r, "DELETE", path, s.Delete)
}
