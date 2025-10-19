package query

import (
	"github.com/go-chi/chi/v5"
)

func (s *Handler) Register(r chi.Router) {
	s.responses.Route(r, "GET", "/metrics/query", s.Instant)
	s.responses.Route(r, "GET", "/metrics/query-range", s.Range)
	s.responses.Route(r, "POST", "/logs/search", s.Logs)
	s.responses.Route(r, "GET", "/traces/services", s.TraceServices)
	s.responses.Route(r, "GET", "/traces/operations", s.TraceOperations)
	s.responses.Route(r, "GET", "/traces/{id}", s.Trace)
	s.responses.Route(r, "POST", "/traces/search", s.SearchTraces)
}
