package query

import (
	"net/http"
	"strings"

	"github.com/Platinaa777/DataCat/internal/api/http/transport"
	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/extapi/telemetry"
	"github.com/Platinaa777/DataCat/internal/pkg/cache"
	"github.com/Platinaa777/DataCat/internal/usecases"
)

type Handler struct {
	service   *usecases.Service
	access    *transport.Access
	responses *transport.Responder
	provider  *telemetry.Provider
	cache     cache.Cache
}

func New(service *usecases.Service, access *transport.Access, responses *transport.Responder, provider *telemetry.Provider, resultCache cache.Cache) *Handler {
	return &Handler{service: service, access: access, responses: responses, provider: provider, cache: resultCache}
}

func (s *Handler) Instant(r *http.Request) (any, error) { return s.cachedMetrics(r, false) }

func (s *Handler) Range(r *http.Request) (any, error) { return s.cachedMetrics(r, true) }

func (s *Handler) Logs(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	input := entity.LogSearch{Page: 1, PageSize: 100}
	if err = transport.DecodeInto(r, &input); err != nil {
		return nil, err
	}
	d, err := s.source(r, ns, input.DataSourceName, "Logs")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(d.Type, "Elasticsearch") {
		return s.plugin(r, d, "logs.search", input)
	}
	out, err := s.provider.Logs(r.Context(), d, input)
	return out, transport.Upstream(err)
}

func (s *Handler) TraceServices(r *http.Request) (any, error) { return s.traceNames(r, "") }

func (s *Handler) TraceOperations(r *http.Request) (any, error) {
	service := transport.QueryValue(r, "serviceName")
	if service == "" {
		return nil, entity.ErrInvalid
	}
	return s.traceNames(r, service)
}

func (s *Handler) Trace(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	d, err := s.source(r, ns, transport.QueryValue(r, "dataSourceName"), "Traces")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(d.Type, "Jaeger") {
		return s.plugin(r, d, "traces.get", map[string]string{"traceId": transport.Param(r, "id")})
	}
	traces, err := s.provider.Traces(r.Context(), d, entity.TraceSearch{}, transport.Param(r, "id"))
	if err != nil {
		return nil, transport.Upstream(err)
	}
	if len(traces) == 0 {
		return nil, entity.ErrNotFound
	}
	return traces[0], nil
}

func (s *Handler) SearchTraces(r *http.Request) (any, error) {
	ns, err := s.access.Scope(r, false)
	if err != nil {
		return nil, err
	}
	input := entity.TraceSearch{Limit: 100}
	if err = transport.DecodeInto(r, &input); err != nil {
		return nil, err
	}
	d, err := s.source(r, ns, input.DataSourceName, "Traces")
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(d.Type, "Jaeger") {
		return s.plugin(r, d, "traces.search", input)
	}
	out, err := s.provider.Traces(r.Context(), d, input, "")
	return out, transport.Upstream(err)
}
