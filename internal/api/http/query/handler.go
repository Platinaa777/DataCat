package query

import (
	"net/http"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Handler struct {
	queries   *usecases.Queries
	access    *transport.Access
	responses *transport.Responder
}

func New(queries *usecases.Queries, access *transport.Access, responses *transport.Responder) *Handler {
	return &Handler{queries: queries, access: access, responses: responses}
}

func (s *Handler) Instant(r *http.Request) (any, error) { return s.metrics(r, false) }

func (s *Handler) Range(r *http.Request) (any, error) { return s.metrics(r, true) }

func (s *Handler) metrics(r *http.Request, isRange bool) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	input := usecases.MetricsQuery{
		DataSource:         transport.QueryValue(r, "dataSourceName"),
		Dashboard:          transport.QueryValue(r, "dashboardId"),
		Query:              transport.QueryValue(r, "query"),
		RequestedNamespace: transport.QueryValue(r, "namespaceId"),
		From:               transport.QueryValue(r, "from"),
		To:                 transport.QueryValue(r, "to"),
		Step:               transport.QueryValue(r, "step"),
	}
	return s.queries.Metrics(r.Context(), ns, input, isRange)
}

func (s *Handler) Logs(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	input := entity.LogSearch{Page: 1, PageSize: 100}
	if err := transport.DecodeInto(r, &input); err != nil {
		return nil, err
	}
	return s.queries.Logs(r.Context(), ns, input)
}

func (s *Handler) TraceServices(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	return s.queries.TraceServices(r.Context(), ns, transport.QueryValue(r, "dataSourceName"))
}

func (s *Handler) TraceOperations(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	return s.queries.TraceOperations(r.Context(), ns, transport.QueryValue(r, "dataSourceName"), transport.QueryValue(r, "serviceName"))
}

func (s *Handler) Trace(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	return s.queries.Trace(r.Context(), ns, transport.QueryValue(r, "dataSourceName"), transport.Param(r, "id"))
}

func (s *Handler) SearchTraces(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	input := entity.TraceSearch{Limit: 100}
	if err := transport.DecodeInto(r, &input); err != nil {
		return nil, err
	}
	return s.queries.SearchTraces(r.Context(), ns, input)
}
